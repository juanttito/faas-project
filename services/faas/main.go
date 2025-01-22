package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
	"github.com/go-redis/redis/v8"
	"github.com/nats-io/nats.go"
	"context"

)

var ctx = context.Background()
var authServiceURL = "http://auth:8000/api/validarusuario" // URL del servicio Auth

type Function struct {
	Usuario string `json:"usuario"`
	Funcion string `json:"funcion"`
	Codigo  string `json:"codigo"`
}

// Registrar una nueva función
func registerFunction(w http.ResponseWriter, r *http.Request) {
	var fn Function

	//Validamos permisos de usuario
	// Obtener el token del encabezado Authorization
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Token requerido", http.StatusUnauthorized)
		return
	}

	// Crear la solicitud HTTP para llamar a /validate en el servicio Auth
	req, err := http.NewRequest("GET", authServiceURL, nil)
	if err != nil {
		http.Error(w, "Error creando la solicitud", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", token)

	// Realizar la solicitud
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Error llamando al servicio Auth", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Verificar la respuesta del servicio Auth
	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Token inválido", http.StatusUnauthorized)
		return
	}

	//fin de validacion

	err = json.NewDecoder(r.Body).Decode(&fn)
	if err != nil {
		http.Error(w, "Carga JSON inválida", http.StatusBadRequest)
		return
	}

	cliente := redis.NewClient(&redis.Options{Addr: "redis:6379"})
	result, err := cliente.HSet(ctx, "functions", fn.Funcion, fn.Codigo).Result()
	if err != nil {
		log.Printf(err.Error())
		http.Error(w, "No se pudo registrar la función en Redis", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"mensaje": "Función registrada con éxito",
		"redis_set": result,
		"nota": "redis_set=1 significa que se añadió una nueva función; redis_set=0 significa que sobrescribió una función existente.",
		"datos": map[string]string{
			"nombre_funcion": fn.Funcion,
			"codigo_funcion": fn.Codigo,
		},
	})
}

// Desregistrar una función existente
func unregisterFunction(w http.ResponseWriter, r *http.Request) {
	var fn Function

		//Validamos permisos de usuario
	// Obtener el token del encabezado Authorization
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Token requerido", http.StatusUnauthorized)
		return
	}

	// Crear la solicitud HTTP para llamar a /validate en el servicio Auth
	req, err := http.NewRequest("GET", authServiceURL, nil)
	if err != nil {
		http.Error(w, "Error creando la solicitud", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", token)

	// Realizar la solicitud
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Error llamando al servicio Auth", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Verificar la respuesta del servicio Auth
	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Token inválido", http.StatusUnauthorized)
		return
	}

	//fin de validacion

	err = json.NewDecoder(r.Body).Decode(&fn)
	if err != nil {
		http.Error(w, "Carga JSON inválida", http.StatusBadRequest)
		return
	}

	cliente := redis.NewClient(&redis.Options{Addr: "redis:6379"})
	result, err := cliente.HDel(ctx, "functions", fn.Funcion).Result()
	if err != nil {
		http.Error(w, "No se pudo desregistrar la función de Redis", http.StatusInternalServerError)
		return
	}

	if result == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"mensaje": "No se encontró ninguna función para desregistrar",
			"redis_del": result,
			"nota": "La función que intentas desregistrar no existe en Redis.",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"mensaje": "Función desregistrada con éxito",
		"redis_del": result,
	})
}

// Llamar a una función registrada
func llamarFunction(w http.ResponseWriter, r *http.Request) {
	var fn Function

	//Validamos permisos de usuario
	// Obtener el token del encabezado Authorization
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Token requerido", http.StatusUnauthorized)
		return
	}

	// Crear la solicitud HTTP para llamar a /validate en el servicio Auth
	req, err := http.NewRequest("GET", authServiceURL, nil)
	if err != nil {
		http.Error(w, "Error creando la solicitud", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", token)

	// Realizar la solicitud
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Error llamando al servicio Auth", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Verificar la respuesta del servicio Auth
	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Token inválido", http.StatusUnauthorized)
		return
	}

	//fin de validacion

	err = json.NewDecoder(r.Body).Decode(&fn)
	if err != nil {
		http.Error(w, "Carga JSON inválida", http.StatusBadRequest)
		return
	}

	cliente := redis.NewClient(&redis.Options{Addr: "redis:6379"})

	// Verificar si la función existe
	exists, err := cliente.HExists(ctx, "functions", fn.Funcion).Result()
	if err != nil || !exists {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Función no encontrada",
			"sugerencia": "Por favor, asegúrate de que la función esté registrada antes de llamarla.",
		})
		return
	}

	// Recuperar el código de la función
	code, err := cliente.HGet(ctx, "functions", fn.Funcion).Result()
	log.Printf(string(code))
	if err != nil {
		http.Error(w, "No se pudo recuperar el código de la función", http.StatusInternalServerError)
		return
	}

	// Conexion a NATS
	nc, err := nats.Connect("nats://nats:4222")
	if err != nil {
		log.Fatalf("Error al conectar a NATS:", err)
		http.Error(w, "Error al conectar a NATS", http.StatusInternalServerError)
		return
	}
	defer nc.Close()

	// Enviar la solicitud a traves de NATS y esperar una respuesta
	msg, err := nc.Request("Peticion", []byte(code), 50*time.Second)
	log.Printf(string(msg.Data))
	if err != nil {
		log.Printf("Error al procesar la solicitud en NATS: %v", err)
		http.Error(w, "Error al procesar la solicitud en NATS", http.StatusInternalServerError)
		return
	}

	// Preparar la respuesta al cliente
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": string(msg.Data)}); err != nil {
		log.Printf("Error al codificar la respuesta JSON: %v", err)
		http.Error(w, "Error al generar la respuesta", http.StatusInternalServerError)
	}

	/*
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"mensaje": "Función ejecutada con éxito",
		"datos": map[string]string{
			"nombre_funcion": fn.Funcion,
			"codigo":          code,
		},
	})
	*/
}

func main() {
	http.HandleFunc("/api/registrafuncion", registerFunction)
	http.HandleFunc("/api/desregistrafuncion", unregisterFunction)
	http.HandleFunc("/api/llamarfuncion", llamarFunction)
	log.Fatal(http.ListenAndServe(":8001", nil))
}