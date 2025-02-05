package recommendation

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"Go-internship-Manifure/internal/model"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
}

// Инициализация нового обработчика рекомендаций.
func NewRecommendationHandler(db *gorm.DB) *Handler {
	return &Handler{DB: db}
}

// Обработчик сообщений для сервиса рекомендаций.
func (rh *Handler) HandleMessage(message []byte, topic kafka.TopicPartition, cn int) error {
	log.Printf("Consumer #%d received message from topic %s: %s", cn, *topic.Topic, string(message))
	// В зависимости от переданного топика kafka, выбирается обработчик
	switch *topic.Topic {
	case "product-updates":
		return rh.HandleProductMessage(message)
	case "user-updates":
		return rh.HandleUserMessage(message)
	default:
		return fmt.Errorf("unknown topic: %s", *topic.Topic)
	}
}

// Обработчик сообщений продукта.
func (rh *Handler) HandleProductMessage(message []byte) error {
	var product model.Recommendations
	if err := json.Unmarshal(message, &product); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	var existingProduct model.Recommendations
	if err := rh.DB.Where("id = ?", product.ID).First(&existingProduct).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			product.PopularityScore = 1
			if err := rh.DB.Create(&product).Error; err != nil {
				return fmt.Errorf("failed to create recommendations: %w", err)
			}

			log.Printf("Recommended product #%s created with ID %s", product.ID, product.ID)
		} else {
			return fmt.Errorf("failed to get recommendations: %w", err)
		}
	} else {
		existingProduct.PopularityScore++
		if err := rh.DB.Save(&existingProduct).Error; err != nil {
			return fmt.Errorf("failed to update recommendations: %w", err)
		}

		log.Printf("Updated product: %+v", existingProduct)
	}

	return nil
}

// Обработчик сообщений пользователя.
func (rh *Handler) HandleUserMessage(message []byte) error {
	var user model.User
	if err := json.Unmarshal(message, &user); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	for _, cartItem := range user.Cart {
		var product model.Recommendations
		if err := rh.DB.Where("id = ?", cartItem.ProductID).First(&product).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				product.ID = cartItem.ProductID
				product.PopularityScore = 1

				if err := rh.DB.Create(&product).Error; err != nil {
					return fmt.Errorf("failed to insert product from user cart: %w", err)
				}

				log.Printf("Product #%s created with ID %s", product.ID, product.ID)
			} else {
				return fmt.Errorf("failed to query product from user cart: %w", err)
			}
		} else {
			product.PopularityScore++
			if err := rh.DB.Save(&product).Error; err != nil {
				return fmt.Errorf("failed to update product from user cart: %w", err)
			}

			log.Printf("Updated product: %+v", product)
		}
	}

	return nil
}
