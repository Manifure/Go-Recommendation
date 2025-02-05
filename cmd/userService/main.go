package main

import (
	"Go-internship-Manifure/internal/auth"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"Go-internship-Manifure/internal/handlers/user"
	k "Go-internship-Manifure/internal/kafka"
	"Go-internship-Manifure/internal/monitoring"
	"github.com/gorilla/mux"
)

const topic = "user-updates"

func main() {
	monitoring.Init()

	kafkaEnv := os.Getenv("KAFKA_ADDRESS")
	if kafkaEnv == "" {
		kafkaEnv = "localhost:9091,localhost:9092,localhost:9093" // Значение по умолчанию
	}

	address := strings.Split(kafkaEnv, ",")

	// Настройка kafka продюсера
	p, err := k.NewProducer(address, topic)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}

	// Инициализация обработчика
	userHandler := user.NewUserHandler(p)

	// Настройка api
	r := mux.NewRouter()
	r.HandleFunc("/users", userHandler.RegisterUser).Methods("POST")
	r.Handle("/users/{id}", auth.JWTMiddleware(http.HandlerFunc(userHandler.GetUser))).Methods("GET")
	r.Handle("/users/{id}", auth.JWTMiddleware(http.HandlerFunc(userHandler.UpdateUser))).Methods("PUT")

	// Подключаем middleware для мониторинга
	r.Use(monitoring.Middleware)

	// Регистрируем обработчик метрик Prometheus
	r.Path("/metrics").Handler(monitoring.MetricsHandler())

	// Настройка HTTP сервера
	serverAddress := ":8080"
	server := &http.Server{
		Addr:    serverAddress,
		Handler: r,
	}

	// Канал для системных сигналов
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	// Запуск HTTP сервера
	go func() {
		log.Printf("Starting server on %s", serverAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %s", err)
		}
	}()

	// Ожидание сигнала завершения
	<-stopChan
	log.Println("Shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	p.Close()

	// Завершение работы HTTP сервера
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Failed to shutdown HTTP server: %v", err)
	}

	log.Println("User service stopped gracefully")
}
