package db

import (
	"fmt"
	"log"

	"Go-internship-Manifure/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DatabaseRecommendationInterface interface {
	CloseRecommendationDB() error
	MigrateRecommendationModels() error
}

type Database struct {
	Conn *gorm.DB
}

// Соединение с базой данных postgres через gorm.
func NewRecommendationDatabase(host, user, password, dbname, port string) *Database {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", host, user, password, dbname, port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	database := &Database{Conn: db}

	// авто миграция, если таблицы не существует
	if err = database.MigrateRecommendationModels(); err != nil {
		log.Fatalf("Failed to migrate models: %v", err)
	}

	log.Println("Successfully connected to database")

	return &Database{Conn: db}
}

// CloseRecommendationDB Close закрывает базу данных.
func (db *Database) CloseRecommendationDB() error {
	log.Println("Closing database connection...")
	sqlDB, err := db.Conn.DB()
	if err != nil {
		return fmt.Errorf("failed to retrieve *sql.DB: %w", err)
	}
	return sqlDB.Close()
}

// MigrateRecommendationModels выполняет миграцию моделей.
func (db *Database) MigrateRecommendationModels() error {
	return db.Conn.AutoMigrate(&model.Recommendations{})
}
