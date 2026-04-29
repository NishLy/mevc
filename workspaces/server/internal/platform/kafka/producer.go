package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(broker string, topic string) *Producer {

	writer := &kafka.Writer{
		Addr:     kafka.TCP(broker),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	return &Producer{
		writer: writer,
	}
}

func (p *Producer) Publish(key string, value []byte) error {

	msg := kafka.Message{
		Key:   []byte(key),
		Value: value,
	}

	return p.writer.WriteMessages(context.Background(), msg)
}
