package handler

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/nats-io/nats.go"
	"log"
	"time"
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
}

func RegisterFunction(fn Function) error {
	return redisClient.HSet(ctx, "functions", fn.Funcion, fn.Usuario).Err()
}

func DeregisterFunction(fn Function) error {
	return redisClient.HDel(ctx, "functions", fn.Funcion).Err()
}

func CallFunction(fn Function) (string, error) {
	fnData, err := json.Marshal(fn)
	if err != nil {
		return "", err
	}
	msg, err := natsConn.Request("Peticion", fnData, 50*time.Second)
	if err != nil {
		return "", err
	}
	return string(msg.Data), nil
}
