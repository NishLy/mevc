package kafka

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(broker, topic, group string) *Consumer {

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic:   topic,
		GroupID: group,
	})

	return &Consumer{reader: reader}
}

func (c *Consumer) Start(handler func([]byte)) {

	for {
		msg, err := c.reader.ReadMessage(context.Background())
		if err != nil {
			log.Println(err)
			continue
		}

		handler(msg.Value)
	}
}
