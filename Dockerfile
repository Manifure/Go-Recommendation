# Устанавливаем базовый образ с Go
FROM golang:1.23 AS builder
LABEL authors="manifure"

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем файлы модулей и загружаем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь код проекта в контейнер
COPY . .

# Устанавливаем переменную для сборки (заменить `cmd/<ServiceName>/main.go`)
ARG SERVICE_PATH
RUN go build -o main $SERVICE_PATH

# Используем минимальный образ для запуска приложения
FROM frolvlad/alpine-glibc

WORKDIR /root/

# Копируем скомпилированный бинарник
COPY --from=builder /app/main .

# Устанавливаем порт для приложения
EXPOSE 8084:8080

# Запускаем приложение, sleep на 30 секунд, что бы необходимые контейнеры успели загрузиться
CMD ["sh", "-c", "sleep 30 && ./main"]