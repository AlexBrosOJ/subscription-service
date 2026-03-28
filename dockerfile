# Этап сборки
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Устанавливаем необходимые зависимости
RUN apk add --no-cache git

# Копируем go.mod и go.sum
COPY go.mod go.sum ./

# Выполняем go mod tidy
RUN go mod tidy

# Скачиваем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -v -o subscription-service ./cmd/app

# Этап выполнения
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем бинарный файл из этапа сборки
COPY --from=builder /app/subscription-service .
COPY --from=builder /app/.env.example .env

EXPOSE 8080

CMD ["./subscription-service"]