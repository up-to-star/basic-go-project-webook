package events

import (
	"basic-project/webook/interactive/repository"
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type InteractiveReadEventConsumer struct {
	reader *kafka.Reader
	repo   repository.InteractiveRepository
}

func (i *InteractiveReadEventConsumer) Start() {
	go func() {
		for {
			msg, err := i.reader.ReadMessage(context.Background())
			if err != nil {
				zap.L().Error("kafka 读取消息失败", zap.Error(err))
				continue
			}
			var evt ReadEvent
			err = json.Unmarshal(msg.Value, &evt)
			if err != nil {
				zap.L().Error("kafka 反序列化消息失败", zap.Error(err))
				continue
			}
			err = i.Consume(context.Background(), evt)
			if err != nil {
				zap.L().Error("kafka 消费消息失败", zap.Error(err))
				continue
			}
		}
	}()
}

func (i *InteractiveReadEventConsumer) Close() error {
	return i.reader.Close()
}

func (i *InteractiveReadEventConsumer) Consume(ctx context.Context, evt ReadEvent) error {
	return i.repo.IncrReadCnt(ctx, "article", evt.Aid)
}
func NewInteractiveReadEventConsumer(addrs []string, repo repository.InteractiveRepository) *InteractiveReadEventConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  addrs,
		GroupID:  "interactive",
		Topic:    "read-article",
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
	return &InteractiveReadEventConsumer{
		reader: reader,
		repo:   repo,
	}
}
