# Dockerfile para un servicio en Go
FROM golang:1.23-alpine
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o auth .
CMD ["./auth"]

