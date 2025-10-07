FROM golang:1.23.2-alpine

WORKDIR /app


RUN apk add --no-cache git


RUN go install github.com/swaggo/swag/cmd/swag@latest


COPY go.mod go.sum ./
RUN go mod download


COPY . .


RUN swag init -g cmd/server/main.go -o docs


RUN go build -o main ./cmd/server


EXPOSE 8080


CMD ["./main"]