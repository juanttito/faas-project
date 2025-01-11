package main

import (
	"fmt"
	"github.com/nats-io/nats.go"
)

func main() {
	nc, _ := nats.Connect("nats://nats:4222")
	defer nc.Close()

	nc.Subscribe("Peticion", func(m *nats.Msg) {
		fmt.Printf("Recibido: %s\n", string(m.Data))
		nc.Publish("Respuesta", []byte("Trabajo completado"))
	})
	select {}
}
