1. Tener instalado Docker, Postman

2. Clonar el proyecto faas de github
	git clone https://github.com/juanttito/faas-project.git
	cd faas-project

3. Compilar el docker-compose.yml , debes esta ubicado en la ruta del archivo
	docker-compose up --build

4. Validamos que los contenedores esten ejecutando
$ docker ps
CONTAINER ID   IMAGE                  COMMAND                  CREATED        STATUS          PORTS                                        NAMES
3877c07a6a7a   trabajos-faas          "./faas"                 22 hours ago   Up 35 minutes   0.0.0.0:8001->8001/tcp                       trabajos-faas-1
fac65f3936a4   trabajos-orquestador   "./orquestador"          22 hours ago   Up 35 minutes   0.0.0.0:8002->8002/tcp                       trabajos-orquestador-1
cf199b558ed2   trabajos-auth          "./auth"                 22 hours ago   Up 35 minutes   0.0.0.0:8000->8000/tcp                       trabajos-auth-1
9c8a99b9a0d1   redis:latest           "docker-entrypoint.s…"   46 hours ago   Up 35 minutes   0.0.0.0:6379->6379/tcp                       trabajos-redis-1
370755333c5b   nats:latest            "/nats-server --conf…"   46 hours ago   Up 35 minutes   6222/tcp, 0.0.0.0:4222->4222/tcp, 8222/tcp   trabajos-nats-1

5. Ingresamos al Postman y importamos el archivo FaaS.postman_collection.json que contiene los endpoint para realziar la pruebas

