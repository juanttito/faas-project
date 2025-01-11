package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-redis/redis/v8"
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

func main() {
	http.HandleFunc("/api/registrafuncion", registerFunction)
	http.HandleFunc("/api/desregistrafuncion", unregisterFunction)
	http.ListenAndServe(":8001", nil)
}
