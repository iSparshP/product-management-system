// internal/infrastructure/kafka/publisher.go

package kafka

import (
	"context"
	"time"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

type Publisher struct {
	producer sarama.SyncProducer
	topic    string
	logger   *zap.Logger
}

func NewPublisher(brokers []string, topic string, logger *zap.Logger) (*Publisher, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Retry.Max = 5
	config.Producer.Retry.Backoff = 100 * time.Millisecond

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &Publisher{
		producer: producer,
		topic:    topic,
		logger:   logger,
	}, nil
}

func (p *Publisher) Publish(ctx context.Context, message []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.ByteEncoder(message),
	}

	_, _, err := p.producer.SendMessage(msg)
	if err != nil {
		p.logger.Error("Failed to send message to Kafka", zap.Error(err))
		return err
	}

	return nil
}
