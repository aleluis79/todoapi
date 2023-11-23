## Todos API - Go

### Actualizar dependencias

go mod tidy

### Reconstruir documentación swagger

swag init -g cmd/todoapi/main.go > /dev/null

### Iniciar aplicación

go run ./cmd/todoapi/main.go