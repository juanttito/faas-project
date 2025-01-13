package main

import (
	"encoding/json"
	"net/http"
	"log"
	"time"
	"github.com/go-redis/redis/v8"
	"github.com/nats-io/nats.go"
	"context"
)

var ctx = context.Background()

type Function struct {
	Usuario string `json:"usuario"`
	Funcion string `json:"funcion"`
}

func registerFunction(w http.ResponseWriter, r *http.Request) {
	var fn Function
	json.NewDecoder(r.Body).Decode(&fn)
	client := redis.NewClient(&redis.Options{Addr: "redis:6379"})
	client.HSet(ctx, "functions", fn.Funcion, fn.Usuario)
	w.WriteHeader(http.StatusOK)
}

func unregisterFunction(w http.ResponseWriter, r *http.Request) {
	var fn Function
	json.NewDecoder(r.Body).Decode(&fn)
	client := redis.NewClient(&redis.Options{Addr: "redis:6379"})
	client.HDel(ctx, "functions", fn.Funcion)
	w.WriteHeader(http.StatusOK)
}
func llamarFunction(w http.ResponseWriter, r *http.Request) {
	var fn Function

	// Decodificar el cuerpo de la solicitud en la estructura Function
	if err := json.NewDecoder(r.Body).Decode(&fn); err != nil {
		http.Error(w, "Error al decodificar JSON: "+err.Error(), http.StatusBadRequest)
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

	// Serializar la estructura Function a JSON
	fnData, err := json.Marshal(fn)
	if err != nil {
		log.Printf("Error al serializar estructura Function: %v", err)
		http.Error(w, "Error al procesar la solicitud", http.StatusInternalServerError)
		return
	}

	// Enviar la solicitud a traves de NATS y esperar una respuesta
	msg, err := nc.Request("Peticion", fnData, 50*time.Second)
	if err != nil {
		log.Printf("Error al procesar la solicitud en NATS: %v", err)
		http.Error(w, "Error al procesar la solicitud en NATS", http.StatusInternalServerError)
		return
	}

	/*json.NewDecoder(r.Body).Decode(&fn)

	msg, err := nc.Request("Peticion", []byte(fn), 50*time.Second)
		if err != nil  {
			fmt.Println("Error al procesar",err)
			
		}

		*/

	// Preparar la respuesta al cliente
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": string(msg.Data)}); err != nil {
		log.Printf("Error al codificar la respuesta JSON: %v", err)
		http.Error(w, "Error al generar la respuesta", http.StatusInternalServerError)
	}
	/*
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": string(msg.Data)})
	*/
}

func main() {
	http.HandleFunc("/api/registrafuncion", registerFunction)
	http.HandleFunc("/api/desregistrafuncion", unregisterFunction)
	http.HandleFunc("/api/llamarfuncion", llamarFunction)
	http.ListenAndServe(":8001", nil)
}
