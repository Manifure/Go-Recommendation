package kafka

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

const (
	flushTimeout = 5000 // ms
)

var errUnknownType = errors.New("unknown event type")

type ProducerInterface interface {
	Produce(msg string) error
}

type Producer struct {
	Producer *kafka.Producer
	Topic    string
}

// Подключается к kafka и создает новый producer.
func NewProducer(addrs []string, topic string) (*Producer, error) {
	conf := &kafka.ConfigMap{
		"bootstrap.servers": strings.Join(addrs, ","),
	}

	p, err := kafka.NewProducer(conf)
	if err != nil {
		return nil, fmt.Errorf("error creating new producer: %w", err)
	}

	return &Producer{
		Producer: p,
		Topic:    topic,
	}, nil
}

// Обработчик сообщений в kafka.
func (p *Producer) Produce(msg string) error {
	kafkaMsg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &p.Topic,
			Partition: kafka.PartitionAny,
		},
		Value:     []byte(msg),
		Key:       nil,
		Timestamp: time.Now(),
	}
	kafkaChan := make(chan kafka.Event)

	if err := p.Producer.Produce(kafkaMsg, kafkaChan); err != nil {
		return fmt.Errorf("error sending message to kafka: %w", err)
	}

	e := <-kafkaChan
	switch ev := e.(type) {
	case *kafka.Message:
		log.Printf("Message produced by Kafka: %s", ev.Value)

		return nil
	case kafka.Error:
		log.Printf("Error produced by Kafka: %s", ev.String())

		return ev
	default:
		return errUnknownType
	}
}

// Закрывает продюсер, после ожидания обработки непринятых сообщений.
func (p *Producer) Close() {
	log.Println("Closing Kafka producer connection...")
	p.Producer.Flush(flushTimeout)
	p.Producer.Close()
}
