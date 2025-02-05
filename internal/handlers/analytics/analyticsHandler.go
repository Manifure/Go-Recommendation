package analytics

import (
	"encoding/json"
	"log"

	"Go-internship-Manifure/internal/model"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type Handler struct {
	DB       *gorm.DB
	Validate *validator.Validate
}

// Создание нового обработчика аналитики.
func NewAnalyticsHandler(db *gorm.DB) *Handler {
	return &Handler{
		DB:       db,
		Validate: validator.New(),
	}
}

// Обработчик сообщений аналитики.
func (h *Handler) HandleMessage(message []byte, topic kafka.TopicPartition, _ int) error {
	switch *topic.Topic {
	case "product-updates":
		log.Printf("Message %s send to %s topic", message, *topic.Topic)
		return h.HandleProductUpdate(message)
	case "user-updates":
		log.Printf("Message %s send to %s topic", message, *topic.Topic)
		return h.HandleUserUpdate(message)
	default:
		log.Printf("Unknown topic: %s", *topic.Topic)

		return nil
	}
}

// Обработчик данных продукта.
func (h *Handler) HandleProductUpdate(message []byte) error {
	var product model.ProductUpdateAnalytics

	// Десериализация сообщения
	if err := json.Unmarshal(message, &product); err != nil {
		log.Printf("Error unmarshalling product update: %v", err)

		return err
	}

	// Валидация данных
	if err := h.Validate.Struct(product); err != nil {
		log.Printf("Error validating product update: %v", err)

		return err
	}

	// Обновление статистики в базе данных
	err := h.DB.Model(&model.ProductStatistics{}).Where("product_id = ?", product.ID).FirstOrCreate(&model.ProductStatistics{ProductID: product.ID}).Update("update_count", gorm.Expr("update_count + ?", 1)).Error
	if err != nil {
		log.Printf("Error updating product statistics: %v", err)
	}

	return err
}

// Обработчик данных пользователя.
func (h *Handler) HandleUserUpdate(message []byte) error {
	var user model.UserUpdateAnalytics

	// Десериализация сообщения
	if err := json.Unmarshal(message, &user); err != nil {
		log.Printf("Error unmarshalling user update: %v", err)

		return err
	}

	// Валидация данных
	if err := h.Validate.Struct(user); err != nil {
		log.Printf("Error validating user update: %v", err)

		return err
	}

	// Обновление статистики в базе данных
	err := h.DB.Model(&model.UserStatistics{}).Where("user_id = ?", user.ID).FirstOrCreate(&model.UserStatistics{UserID: user.ID}).Update("activity_count", gorm.Expr("activity_count + ?", 1)).Error
	if err != nil {
		log.Printf("Error updating user activity count: %v", err)
	}

	return err
}
