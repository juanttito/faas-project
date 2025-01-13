package main

import (
	"fmt"
	"github.com/nats-io/nats.go"
)

func main() {
	nc, _ := nats.Connect("nats://nats:4222")
	defer nc.Close()

	nc.Subscribe("Respuesta", func(m *nats.Msg) {
		fmt.Printf("Recibido: %s\n", string(m.Data))
		m.Respond([]byte("Imagen Ejecutado desde worker"))
	})
	fmt.Printf("Se Ejecuto la imagen del worker")

}
