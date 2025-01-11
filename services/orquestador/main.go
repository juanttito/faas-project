package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/nats-io/nats.go"
)

type Trabajo struct {
	Funcion   string `json:"funcion"`
	Parametro string `json:"parametro"`
}

func main() {

	nc, err := nats.Connect("nats://nats:4222")
	if err != nil {
		log.Fatalf("Error al conectar a NATS: %v", err)
	}
	defer nc.Close()

	nc.Subscribe("Respuesta", func(m *nats.Msg) {
		fmt.Printf("Respuesta: %s\n", string(m.Data))
		//nc.Publish("Respuesta", []byte("Trabajo completado"))
	})

	// Endpoint para registrar trabajo
	http.HandleFunc("/api/registrartrabajo", func(w http.ResponseWriter, r *http.Request) {
		// Verifica que el método sea POST
		if r.Method != http.MethodPost {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
			return
		}

		// Leer el cuerpo de la solicitud
		var trabajo Trabajo
		err := json.NewDecoder(r.Body).Decode(&trabajo)
		if err != nil {
			http.Error(w, "Error al leer el cuerpo de la solicitud", http.StatusBadRequest)
			return
		}

		// Crear el mensaje con el trabajo recibido
		msg := fmt.Sprintf("Funcion: %s, Parametro: %s", trabajo.Funcion, trabajo.Parametro)

		// Publicar el mensaje en el tópico "Peticion" de NATS
		err = nc.Publish("Peticion", []byte(msg))
		if err != nil {
			http.Error(w, "Error al enviar mensaje a NATS", http.StatusInternalServerError)
			return
		}

		// Responder al cliente
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Trabajo registrado y mensaje enviado a NATS"))
	})

	// Iniciar el servidor HTTP
	log.Println("Servidor en ejecución en http://localhost:8002")
	log.Fatal(http.ListenAndServe(":8002", nil))
}
