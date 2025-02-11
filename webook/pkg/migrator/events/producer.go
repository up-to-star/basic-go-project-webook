package events

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
)

type Producer interface {
	ProduceInconsistentEvent(ctx context.Context, evt InconsistentEvent) error
}

type KafkaProducer struct {
	producer *kafka.Writer
	topic    string
}

func (k *KafkaProducer) ProduceInconsistentEvent(ctx context.Context, evt InconsistentEvent) error {
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	return k.producer.WriteMessages(ctx, kafka.Message{
		Topic: k.topic,
		Value: data,
	})
}

func NewKafkaProducer(addrs []string, topic string) Producer {
	return &KafkaProducer{
		producer: &kafka.Writer{
			Addr:     kafka.TCP(addrs...),
			Balancer: &kafka.LeastBytes{},
		},
		topic: topic,
	}
}
