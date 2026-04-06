package kafka

import (
	"context"
	"errors"
	"time"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer() (*Producer, error) {
	zap.L().Info("开始初始化Kafka生产者",
		zap.String("component", "kafka"),
	)

	writer := &kafka.Writer{
		Addr:         kafka.TCP("localhost:9092"),
		Topic:        "user-events",
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Async:        false,
	}

	conn, err := kafka.Dial("tcp", "localhost:9092")
	if err != nil {
		zap.L().Error("× Kafka连接失败",
			zap.Error(err),
			zap.String("component", "kafka"),
		)
		return nil, errors.New("Kafka连接失败: " + err.Error())
	}
	conn.Close()

	zap.L().Info("√ Kafka生产者初始化成功",
		zap.String("component", "kafka"),
	)

	return &Producer{writer: writer}, nil
}

func (p *Producer) SendUserEvent(ctx context.Context, key string, value []byte) error {
	tracer := otel.Tracer("kafka")
	_, span := tracer.Start(ctx, "kafka.SendUserEvent")
	defer span.End()

	span.SetAttributes(
		attribute.String("kafka.operation", "produce"),
		attribute.String("kafka.topic", "user-events"),
		attribute.String("kafka.key", key),
	)

	err := p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: value,
		Time:  time.Now(),
	})

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		zap.L().Error("× 发送Kafka消息失败",
			zap.Error(err),
			zap.String("key", key),
			zap.String("component", "kafka"),
		)
		return err
	}

	zap.L().Info("√ 发送Kafka消息成功",
		zap.String("key", key),
		zap.String("component", "kafka"),
	)

	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(topic string, groupID string) (*Consumer, error) {
	zap.L().Info("开始初始化Kafka消费者",
		zap.String("topic", topic),
		zap.String("group_id", groupID),
		zap.String("component", "kafka"),
	)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"localhost:9092"},
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	zap.L().Info("√ Kafka消费者初始化成功",
		zap.String("component", "kafka"),
	)

	return &Consumer{reader: reader}, nil
}

func (c *Consumer) Consume(ctx context.Context, handler func(key, value []byte) error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				zap.L().Error("× 读取Kafka消息失败",
					zap.Error(err),
					zap.String("component", "kafka"),
				)
				continue
			}

			if err := handler(msg.Key, msg.Value); err != nil {
				zap.L().Error("× 处理Kafka消息失败",
					zap.Error(err),
					zap.String("component", "kafka"),
				)
			}
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
