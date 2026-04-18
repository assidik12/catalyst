package event

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(writer *kafka.Writer) *KafkaProducer {
	return &KafkaProducer{writer: writer}
}

func (k *KafkaProducer) Publish(ctx context.Context, topic string, data any) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return k.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Value: payload,
	})
}
