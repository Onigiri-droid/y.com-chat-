# Используем образ для Go
FROM golang:1.22 AS builder
WORKDIR /app

# Копируем все файлы проекта и загружаем зависимости
COPY . .
RUN go mod download

# Компиляция проекта
RUN go build -o main .

# Минимальный образ для запуска
FROM debian:latest
RUN apt-get update && apt-get install -y netcat-openbsd
WORKDIR /app
COPY --from=builder /app/main .
COPY wait-for-it.sh /app/wait-for-it.sh

# Делаем wait-for-it.sh исполняемым
RUN chmod +x /app/wait-for-it.sh

# Открываем порт
EXPOSE 50053

# Команда для запуска приложения
CMD ["./wait-for-it.sh", "api-service", "8080", "--", "./main"]

# # Открываем порт
# EXPOSE 8081 50052

# # Команда для запуска приложения
# CMD ["./wait-for-it.sh", "api-service:50051", "--", "./wait-for-it.sh", "mongo-db:27017", "--", "./main"]