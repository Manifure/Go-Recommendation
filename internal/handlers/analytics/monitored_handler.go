package analytics

import (
	"log"
	"time"

	"Go-internship-Manifure/internal/monitoring"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// Добавляет сбор метрик в обработчике сообщений.
type MonitoredHandler struct {
	handler Handler
}

func NewMonitoredHandler(handler *Handler) *MonitoredHandler {
	return &MonitoredHandler{handler: *handler}
}

// Дорабатывает сообщение Kafka с мониторингом.
func (mh *MonitoredHandler) HandleMessage(message []byte, topic kafka.TopicPartition, cn int) error {
	start := time.Now()
	err := mh.handler.HandleMessage(message, topic, cn)
	duration := time.Since(start).Seconds()

	topicName := *topic.Topic
	status := "success"

	if err != nil {
		status = "error"

		log.Printf("Error handling message from topic %s: %v", topicName, err)
	}

	// Обновление метрик
	monitoring.KafkaMessagesConsumedTotal.WithLabelValues(topicName, status).Inc()
	monitoring.KafkaMessageProcessingDuration.WithLabelValues(topicName).Observe(duration)

	return err
}
