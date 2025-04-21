package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/sdshorin/generia/pkg/logger"
	"go.uber.org/zap"
)

// Producer is a Kafka producer client
type Producer struct {
	writer *kafka.Writer
}

// NewProducer creates a new Kafka producer
func NewProducer(brokers []string) *Producer {
	// Create a writer with brokers
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
	}

	return &Producer{
		writer: writer,
	}
}

// Send sends a message to a topic
func (p *Producer) Send(topic string, message []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Value: message,
	})

	if err != nil {
		logger.Logger.Error("Failed to send message to Kafka",
			zap.Error(err),
			zap.String("topic", topic))
		return fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	return nil
}

// SendJSON marshals an object to JSON and sends it to a topic
func (p *Producer) SendJSON(topic string, data interface{}) error {
	messageJSON, err := json.Marshal(data)
	if err != nil {
		logger.Logger.Error("Failed to marshal message to JSON", zap.Error(err))
		return fmt.Errorf("failed to marshal message to JSON: %w", err)
	}

	return p.Send(topic, messageJSON)
}

// Close closes the producer
func (p *Producer) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}