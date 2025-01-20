package main

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"log"
)

func main() {
	// Conectar al servidor NATS
	nc, err := nats.Connect("nats://nats:4222")
	if err != nil {
		log.Fatalf("Error al conectar con NATS: %v", err)
	}
	defer nc.Close()

	// Suscribirse al tema "Respuesta"
	_, err = nc.Subscribe("Respuesta", func(m *nats.Msg) {
		fmt.Printf("Recibido: %s\n", string(m.Data))
		// Responder al remitente del mensaje
		m.Respond([]byte("Imagen Ejecutado worker1"))
	})
	if err != nil {
		log.Fatalf("Error al suscribirse al tema de NATS: %v", err)
	}

	// Imprimir información de inicio
	fmt.Println("Worker está ejecutando y escuchando mensajes en el tema 'Respuesta'...")

	// Evitar que el hilo principal termine
	select {}
}
