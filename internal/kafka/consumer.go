package kafka

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
	"strings"
)

const (
	sessionTimeout     = 7000 // ms
	autoCommitInterval = 5000
	readTimeout        = 10000
)

type Handler interface {
	HandleMessage(message []byte, topic kafka.TopicPartition, cn int) error
}

type Consumer struct {
	Consumer       *kafka.Consumer
	Handler        Handler
	consumerNumber int
}

// Подключается к kafka и инициализирует новый consumer.
func NewConsumer(handler Handler, address, topic []string, consumerGroup string, consumerNumber int) (*Consumer, error) {
	cfg := &kafka.ConfigMap{
		"bootstrap.servers":        strings.Join(address, ","),
		"group.id":                 consumerGroup,
		"session.timeout.ms":       sessionTimeout,
		"enable.auto.offset.store": false,
		"enable.auto.commit":       true,
		"auto.commit.interval.ms":  autoCommitInterval,
		"auto.offset.reset":        "earliest",
	}

	consumer, err := kafka.NewConsumer(cfg)
	if err != nil {
		return nil, err
	}

	err = consumer.SubscribeTopics(topic, nil)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		Consumer:       consumer,
		Handler:        handler,
		consumerNumber: consumerNumber,
	}, nil
}

// Запуск обработчика сообщений kafka.
func (c *Consumer) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Kafka consumer...")

			return
		default:
			kafkaMsg, err := c.Consumer.ReadMessage(readTimeout)
			if err != nil {
				if err.(kafka.Error).Code() == kafka.ErrTimedOut {
					continue
				}
				log.Printf("error reading message from kafka: %v", err)
				continue
			}

			if kafkaMsg == nil {
				continue
			}

			// Обработка сообщения
			if err = c.Handler.HandleMessage(kafkaMsg.Value, kafkaMsg.TopicPartition, c.consumerNumber); err != nil {
				log.Printf("error handle message: %v", err)

				continue
			}

			// Фиксация смещения сообщения
			if _, err = c.Consumer.StoreMessage(kafkaMsg); err != nil {
				log.Printf("error store message: %v", err)

				continue
			}
		}
	}
}

// Закрывает consumer.
func (c *Consumer) Close() error {
	log.Println("Closing Kafka consumer connection...")

	return c.Consumer.Close()
}
