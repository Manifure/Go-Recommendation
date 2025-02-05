package main

import (
	"Go-internship-Manifure/internal/db/analytics_db"
	"Go-internship-Manifure/internal/handlers/analytics"
	"Go-internship-Manifure/internal/kafka"
	"Go-internship-Manifure/internal/monitoring"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	monitoring.Init()

	kafkaEnv := os.Getenv("KAFKA_ADDRESS")
	if kafkaEnv == "" {
		kafkaEnv = "localhost:9091,localhost:9092,localhost:9093" // Значение по умолчанию
	}

	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "localhost" // Значение по умолчанию
	}

	user := os.Getenv("POSTGRES_USER")
	if user == "" {
		user = "postgres"
	}

	password := os.Getenv("POSTGRES_PASSWORD")
	if password == "" {
		password = "1"
	}

	dbname := os.Getenv("POSTGRES_DB")
	if dbname == "" {
		dbname = "postgres"
	}

	port := os.Getenv("POSTGRES_PORT")
	if port == "" {
		port = "5432"
	}

	address := strings.Split(kafkaEnv, ",")
	consumerGroup := "analytics_service"
	topics := []string{"product-updates", "user-updates"}

	// Соединение с базой данных
	database := db.NewAnalyticsDatabase(host, user, password, dbname, port)

	// Инициализация обработчика
	analyticsHandler := analytics.NewAnalyticsHandler(database.Conn)

	monitoredHandler := analytics.NewMonitoredHandler(analyticsHandler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Канал для системных сигналов
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	// настройка kafka консьюмера
	consumer, err := kafka.NewConsumer(monitoredHandler, address, topics, consumerGroup, 1)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}

	// Запуск kafka consumer в отдельной горутине
	go func() {
		log.Println("Starting kafka consumer")
		consumer.Start(ctx)
	}()

	// HTTP сервер для метрик
	metricsServer := &http.Server{
		Addr:    ":8083",
		Handler: monitoring.MetricsHandler(),
	}

	// Запуск HTTP сервера для метрик
	go func() {
		log.Println("Starting metrics server on :8083")
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start metrics server: %v", err)
		}
	}()

	// Ожидание сигнала завершения
	<-stopChan
	log.Println("Shutting down gracefully...")

	cancel()

	// Завершение работы Kafka consumer
	if err := consumer.Close(); err != nil {
		log.Fatalf("Error closing kafka consumer: %v", err)
	}

	// Завершение работы базы данных
	if err := database.CloseAnalyticsDB(); err != nil {
		log.Fatalf("Error closing database connection: %v", err)
	}

	// Завершение работы HTTP сервера
	log.Println("Shutting down metrics server...")

	if err := metricsServer.Shutdown(ctx); err != nil {
		log.Fatalf("Failed to shutdown metrics server: %v", err)
	}

	log.Println("Analytics service stopped gracefully")
}
