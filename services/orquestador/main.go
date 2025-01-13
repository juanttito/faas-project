package main

import (
	"encoding/json"
	"context"
	"fmt"
	"log"
	"time"
	"os"
	"io"
	"bytes"
	"github.com/nats-io/nats.go"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)
var cli *client.Client
var err error
// Estructura para mapear el JSON recibido
type Message struct {
	Funcion   string `json:"funcion"`
	Parametro string `json:"parametro"`
}

func main() {
	//ctx := context.Background()
	// Configurar cliente Docker
	//var err error
	cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Error creando cliente Docker: %v", err)
	}
	// Conexión a NATS
	nc, err := nats.Connect("nats://nats:4222")
	if err != nil {
		log.Fatalf("Error al conectar a NATS: %v", err)
	}
	defer nc.Close()

	// Monitorizar mensajes de trabajo
	go monitorJobs(nc)
	fmt.Println("Servicio en ejecución...")
	select {} // Mantener el servicio en ejecución


}
func monitorJobs(nc *nats.Conn) {
	//for {
		/*msg, err := nc.Request("Peticion", []byte("¿Hay trabajo?"), 2*time.Second)
		if err == nil && string(msg.Data) == "Trabajo disponible" {
			fmt.Println("Trabajo detectado, creando un nuevo worker...")
			err := createWorker()
			if err != nil {
				log.Printf("Error creando Worker: %v", err)
			}
		}*/
		// Imagen de Docker y comando a ejecutar
	//imageName := "busybox"
	command := []string{"echo", "Hola desde Docker con Go!"}

	nc.Subscribe("Peticion", func(m *nats.Msg) {

		var mensaje Message

		// Deserializar el mensaje JSON
		err := json.Unmarshal(m.Data, &mensaje)
		if err != nil {
			log.Printf("Error deserializando mensaje JSON: %v", err)
			return
		}
		
		fmt.Printf("Recibido: %s\n", string(m.Data))
		stdout, err := createWorker(mensaje.Funcion, command)
		if err != nil {
			log.Fatalf("Error gestionando contenedor: %v", err)
		}
		fmt.Printf(stdout)
		m.Respond([]byte(stdout))
			
	})

	select {}

}

// Gestiona un contenedor basado en la imagen de Docker proporcionada y devuelve el stdout
func createWorker(imageName string, command []string) (string, error) {
	ctx := context.Background()
	containerName := fmt.Sprintf("worker-%d", time.Now().UnixNano())

	// Crear cliente Docker
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", fmt.Errorf("error creando cliente Docker: %w", err)
	}

	// Verificar si la imagen existe localmente
	fmt.Println("Verificando imagen:", imageName)
	_, _, err = cli.ImageInspectWithRaw(ctx, imageName)
	if err != nil {
		// Descargar la imagen si no existe
		fmt.Println("Imagen no encontrada localmente. Descargando:", imageName)
		out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
		if err != nil {
			return "", fmt.Errorf("error descargando imagen: %w", err)
		}
		defer out.Close()
		io.Copy(os.Stdout, out)
	}

	// Crear el contenedor
	fmt.Printf("Creando Worker: %s\n", containerName)
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName, // Imagen Docker de tu worker
		//Cmd:   command,
		Tty:   false,
	}, nil, nil, nil, containerName)
	if err != nil {
		return "", fmt.Errorf("error creando contenedor: %w", err)
	}

	// Iniciar el contenedor
	fmt.Println("Iniciando Worker...")
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", fmt.Errorf("error iniciando contenedor: %w", err)
	}

	// Esperar a que el contenedor termine
	fmt.Println("Esperando finalización del contenedor...")
	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return "", fmt.Errorf("error esperando contenedor: %w", err)
		}
	case <-statusCh:
	}
	// Capturar los logs del contenedor
	fmt.Println("Capturando logs del contenedor...")
	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return "", fmt.Errorf("error obteniendo logs del contenedor: %w", err)
	}
	defer out.Close()
	//io.Copy(os.Stdout, out)

	// Leer logs en un buffer
	var stdoutBuf bytes.Buffer
	if _, err := io.Copy(&stdoutBuf, out); err != nil {
		return "", fmt.Errorf("error leyendo logs del contenedor: %w", err)
	}

	// Limpiar el contenedor
	fmt.Println("\nLimpiando contenedor...")
	if err := cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{Force: true}); err != nil {
		return "", fmt.Errorf("error eliminando contenedor: %w", err)
	}

	fmt.Println("Ejecución completada exitosamente.")

	return stdoutBuf.String(), nil
}
