package main

import (
	"encoding/json"
	//"fmt"
	"net/http"

	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v4"
	"context"
)

var ctx = context.Background()

type User struct {
	Correo string `json:"correo"`
	Pswd   string `json:"pswd"`
}

var jwtKey = []byte("your_secret_key")

func registerUser(w http.ResponseWriter, r *http.Request) {
	var user User
	json.NewDecoder(r.Body).Decode(&user)
	client := redis.NewClient(&redis.Options{Addr: "redis:6379"})

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Pswd), bcrypt.DefaultCost)
	client.Set(ctx, user.Correo, hashedPassword, 0)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"correo": user.Correo})
	tokenString, _ := token.SignedString(jwtKey)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func validateUser(w http.ResponseWriter, r *http.Request) {
	tokenString := r.Header.Get("Authorization")
	token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if _, ok := token.Claims.(jwt.MapClaims); !ok || !token.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func main() {
	http.HandleFunc("/api/registro", registerUser)
	http.HandleFunc("/api/validarusuario", validateUser)
	http.ListenAndServe(":8000", nil)
}
