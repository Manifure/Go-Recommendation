package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"Go-internship-Manifure/internal/db/recommendation_db"
	"Go-internship-Manifure/internal/handlers/recommendation"
	"Go-internship-Manifure/internal/kafka"
	"Go-internship-Manifure/internal/monitoring"
	"Go-internship-Manifure/internal/redis"
	"github.com/gorilla/mux"
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
		user = "postgres" // Значение по умолчанию
	}

	password := os.Getenv("POSTGRES_PASSWORD")
	if password == "" {
		password = "1" // Значение по умолчанию
	}

	dbname := os.Getenv("POSTGRES_DB")
	if dbname == "" {
		dbname = "postgres" // Значение по умолчанию
	}

	port := os.Getenv("POSTGRES_PORT")
	if port == "" {
		port = "5432" // Значение по умолчанию
	}

	redisAddress := os.Getenv("REDIS_ADDRESS")
	if redisAddress == "" {
		redisAddress = "localhost:6379" // Значение по умолчанию
	}

	address := strings.Split(kafkaEnv, ",")
	consumerGroup := "recommendation_service"
	topics := []string{"product-updates", "user-updates"}

	// Подключение к базе данных
	database := db.NewRecommendationDatabase(host, user, password, dbname, port)

	// Подключение к redis
	cache := redis.NewCache(redisAddress, "", 0)

	// Инициализация обработчиков
	recommendationHandler := recommendation.NewRecommendationHandler(database.Conn)

	apiHandler := recommendation.NewRecommendationAPIHandler(database.Conn, cache)

	// Контекст для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Канал для системных сигналов
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	// Настройка kafka консьюмера
	consumer, err := kafka.NewConsumer(recommendationHandler, address, topics, consumerGroup, 1)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}

	// Запуск kafka consumer в отдельной горутине
	go func() {
		log.Println("Starting Kafka consumer...")

		consumer.Start(ctx)
	}()

	// Настройка api
	r := mux.NewRouter()
	r.HandleFunc("/recommendations", apiHandler.GetRecommendations).Methods("GET")

	r.Use(monitoring.Middleware)

	// Эндпоинт для метрик
	r.Path("/metrics").Handler(monitoring.MetricsHandler())

	// Запуск HTTP сервера
	serverAddress := ":8082"
	server := &http.Server{
		Addr:    serverAddress,
		Handler: r,
	}

	go func() {
		log.Printf("Starting HTTP server on %s", serverAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to listen on %s: %v", serverAddress, err)
		}
	}()

	// Ожидание сигнала завершения
	<-stopChan
	log.Println("Shutting down gracefully...")

	cancel()

	// Завершение работы Kafka consumer
	if err := consumer.Close(); err != nil {
		log.Fatalf("Error closing consumer: %v", err)
	}

	// Завершение работы базы данных
	if err := database.CloseRecommendationDB(); err != nil {
		log.Fatalf("Error closing database connection: %v", err)
	}

	// Завершение работы Redis
	if err := cache.Close(); err != nil {
		log.Fatalf("Error closing redis connection: %v", err)
	}

	// Завершение работы HTTP сервера
	log.Println("Shutting down metrics server...")

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Error shutting down server: %v", err)
	}

	log.Println("Recommendation service stopped gracefully")
}
