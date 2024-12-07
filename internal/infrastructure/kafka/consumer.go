// internal/infrastructure/kafka/consumer.go

package kafka

import (
	"context"
	"encoding/json"

	"github.com/IBM/sarama"
	"github.com/iSparshP/product-management-system/internal/domain/model"
	"github.com/iSparshP/product-management-system/internal/domain/processor"
	"go.uber.org/zap"
)

type Consumer struct {
	consumerGroup sarama.ConsumerGroup
	topic         string
	logger        *zap.Logger
}

func NewConsumer(brokers []string, groupID string, topic string, logger *zap.Logger) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		consumerGroup: consumerGroup,
		topic:         topic,
		logger:        logger,
	}, nil
}

func (c *Consumer) Close() error {
	return c.consumerGroup.Close()
}

type consumerGroupHandler struct {
	processor processor.ImageProcessor
	logger    *zap.Logger
}

func (c *Consumer) Start(ctx context.Context, processor processor.ImageProcessor) error {
	handler := &consumerGroupHandler{
		processor: processor,
		logger:    c.logger,
	}

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Context cancelled, stopping consumer")
			return ctx.Err()
		default:
			if err := c.consumerGroup.Consume(ctx, []string{c.topic}, handler); err != nil {
				if err == sarama.ErrClosedConsumerGroup {
					c.logger.Info("Consumer group closed")
					return nil
				}
				c.logger.Error("Error from consumer", zap.Error(err))
				return err
			}
		}
	}
}

func (h *consumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		select {
		case <-session.Context().Done():
			return nil
		default:
			var task model.ImageProcessingTask
			if err := json.Unmarshal(message.Value, &task); err != nil {
				h.logger.Error("Failed to unmarshal message",
					zap.Error(err),
					zap.Binary("message_value", message.Value),
					zap.String("topic", message.Topic),
					zap.Int32("partition", message.Partition),
					zap.Int64("offset", message.Offset))
				session.MarkMessage(message, "")
				continue
			}

			// Process the image
			if err := h.processor.ProcessImageTask(task); err != nil {
				h.logger.Error("Failed to process image task",
					zap.Error(err),
					zap.String("product_id", task.ProductID),
					zap.Strings("image_urls", task.ImageURLs))
				// TODO: Implement retry logic or send to DLQ
				// For now, we'll mark the message as processed to avoid infinite loops
			}

			session.MarkMessage(message, "")
		}
	}
	return nil
}
