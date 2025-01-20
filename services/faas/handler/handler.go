package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"bytes"
	"github.com/nats-io/nats.go"
	"github.com/go-redis/redis/v8"
	"log"
)

var ctx = context.Background()
var redisClient *redis.Client
var natsConn *nats.Conn

func InitConnections() {
	redisClient = redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})
	var err error
	natsConn, err = nats.Connect("nats://nats:4222")
	if err != nil {
		log.Fatalf("Error connecting to NATS: %v", err)
	}
}

type Function struct {
	Usuario string `json:"usuario"`
	Funcion string `json:"funcion"`
	Codigo  string `json:"codigo"`
}

// Verifica la autenticación mediante un token
func checkAuth(token string) bool {
	authURL := "http://auth:8000/api/validarusuario"
	req, _ := http.NewRequest("GET", authURL, nil)
	req.Header.Set("Authorization", token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return false
	}
	return true
}

// Valida si una función existe en Redis
func validateFunctionExistence(funcName string) bool {
	result, err := redisClient.Exists(ctx, funcName).Result()
	if err != nil {
		log.Printf("Error en Redis: %v", err)
		return false
	}
	return result > 0
}

// Registra una nueva función
func RegisterFunction(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if !checkAuth(token) {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}
	var fn Function
	if err := json.NewDecoder(r.Body).Decode(&fn); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	if fn.Usuario == "" || fn.Funcion == "" || fn.Codigo == "" {
		http.Error(w, "Faltan campos requeridos", http.StatusBadRequest)
		return
	}
	if validateFunctionExistence(fn.Funcion) {
		http.Error(w, "La función ya está registrada", http.StatusConflict)
		return
	}
	faasURL := "http://faas:8001/api/registrafuncion"
	jsonData, _ := json.Marshal(fn)
	req, _ := http.NewRequest("POST", faasURL, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		http.Error(w, "Error en la solicitud a FaaS", http.StatusInternalServerError)
		return
	}
	if err := redisClient.Set(ctx, fn.Funcion, fn.Codigo, 0).Err(); err != nil {
		http.Error(w, "Error al guardar la función en Redis", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"mensaje": "Función registrada con éxito"})
}

// Llama a una función registrada
func CallFunction(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if !checkAuth(token) {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}
	var fn Function
	if err := json.NewDecoder(r.Body).Decode(&fn); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	if fn.Funcion == "" {
		http.Error(w, "Nombre de la función faltante", http.StatusBadRequest)
		return
	}
	if !validateFunctionExistence(fn.Funcion) {
		http.Error(w, "Función no encontrada", http.StatusNotFound)
		return
	}

	// Inicia un nuevo contenedor Worker
	workerName := "worker_" + fn.Funcion
	cmd := exec.Command("docker", "run", "--rm", "--name", workerName, "worker:latest")
	err := cmd.Start()
	if err != nil {
		http.Error(w, "Error al iniciar el Worker", http.StatusInternalServerError)
		log.Printf("Error al iniciar el Worker: %v", err)
		return
	}

	// Publica un mensaje en el tema "Respuesta" de NATS
	nc, _ := nats.Connect("nats://nats:4222")
	defer nc.Close()
	msg := fmt.Sprintf("Llamando a la función: %s", fn.Funcion)
	nc.Publish("Respuesta", []byte(msg))

	// Espera a que el Worker termine (opcional)
	err = cmd.Wait()
	if err != nil {
		log.Printf("Error al ejecutar el Worker: %v", err)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"mensaje": "Función ejecutada con éxito"})
}

// Desregistra una función existente
func DeregisterFunction(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if !checkAuth(token) {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}
	var fn Function
	if err := json.NewDecoder(r.Body).Decode(&fn); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	if fn.Funcion == "" {
		http.Error(w, "Nombre de la función faltante", http.StatusBadRequest)
		return
	}
	if !validateFunctionExistence(fn.Funcion) {
		http.Error(w, "Función no encontrada", http.StatusNotFound)
		return
	}
	faasURL := "http://faas:8001/api/desregistrafuncion"
	jsonData, _ := json.Marshal(fn)
	req, _ := http.NewRequest("POST", faasURL, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		http.Error(w, "Error en la solicitud a FaaS", http.StatusInternalServerError)
		return
	}
	if err := redisClient.Del(ctx, fn.Funcion).Err(); err != nil {
		http.Error(w, "Error al eliminar la función en Redis", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"mensaje": "Función eliminada con éxito"})
}
