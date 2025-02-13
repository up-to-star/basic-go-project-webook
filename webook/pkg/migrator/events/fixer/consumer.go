package fixer

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/basic-go-project-webook/webook/pkg/migrator"
	"github.com/basic-go-project-webook/webook/pkg/migrator/events"
	"github.com/basic-go-project-webook/webook/pkg/migrator/fixer"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Consumer[T migrator.Entity] struct {
	reader   *kafka.Reader
	srcFirst *fixer.OverrideFixer[T]
	dstFirst *fixer.OverrideFixer[T]
}

func NewConsumer[T migrator.Entity](addrs []string, topic string, src *gorm.DB, dst *gorm.DB) (*Consumer[T], error) {
	srcFirst, err := fixer.NewOverrideFixer[T](src, dst)
	if err != nil {
		return nil, err
	}
	dstFirst, err := fixer.NewOverrideFixer[T](dst, src)
	if err != nil {
		return nil, err
	}
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  addrs,
		Topic:    topic,
		GroupID:  "migrator-fixer",
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
	return &Consumer[T]{
		reader:   reader,
		srcFirst: srcFirst,
		dstFirst: dstFirst,
	}, nil
}

func (c *Consumer[T]) Start() {
	go func() {
		for {
			msg, err := c.reader.ReadMessage(context.Background())
			if err != nil {
				zap.L().Error("kafka 读取消息失败", zap.Error(err))
				continue
			}
			var evt events.InconsistentEvent
			if err := json.Unmarshal(msg.Value, &evt); err != nil {
				zap.L().Error("kafka 反序列化消息失败", zap.Error(err))
				continue
			}
			err = c.Consume(context.Background(), evt)
			if err != nil {
				zap.L().Error("kafka 消费消息失败", zap.Error(err))
				continue
			}
		}
	}()
}

func (c *Consumer[T]) Consume(ctx context.Context, evt events.InconsistentEvent) error {
	switch evt.Direction {
	case "SRC":
		return c.srcFirst.Fix(ctx, evt.ID)
	case "DST":
		return c.dstFirst.Fix(ctx, evt.ID)
	}
	return errors.New("未知的方向")
}

func (c *Consumer[T]) Close() error {
	return c.reader.Close()
}
