package article

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
)

type Producer interface {
	ProduceReadEvent(ctx context.Context, evt ReadEvent) error
}

type KafkaProducer struct {
	producer *kafka.Writer
}

func (k *KafkaProducer) ProduceReadEvent(ctx context.Context, evt ReadEvent) error {
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	return k.producer.WriteMessages(ctx, kafka.Message{
		Topic: "read-article",
		Value: data,
	})
}

func NewKafkaProducer(addrs []string) Producer {
	return &KafkaProducer{
		producer: &kafka.Writer{
			Addr:     kafka.TCP(addrs...),
			Balancer: &kafka.LeastBytes{},
		},
	}
}

type ReadEvent struct {
	Uid int64
	Aid int64
}
