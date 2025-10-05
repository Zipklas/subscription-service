FROM golang:1.23.2-alpine

WORKDIR /app

# Устанавливаем git и зависимости для swag
RUN apk add --no-cache git

# Устанавливаем swag для генерации документации
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Копируем файлы модулей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Генерируем Swagger документацию
RUN swag init -g cmd/server/main.go -o docs

# Собираем приложение
RUN go build -o main ./cmd/server

# Экспонируем порт
EXPOSE 8080

# Запускаем приложение
CMD ["./main"]