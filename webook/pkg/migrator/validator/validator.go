package validator

import (
	"context"
	"errors"
	"github.com/basic-go-project-webook/webook/pkg/migrator"
	"github.com/basic-go-project-webook/webook/pkg/migrator/events"
	"github.com/ecodeclub/ekit/slice"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"log"
	"time"
)

type Validator[T migrator.Entity] struct {
	base      *gorm.DB
	target    *gorm.DB
	direction string
	batchSize int
	fromBase  func(ctx context.Context, offset int) (T, error)
	// <= 0 就认为中断
	// > 0 就认为睡眠
	sleepInterval time.Duration
	p             events.Producer
	utime         int64
}

func NewValidator[T migrator.Entity](base *gorm.DB, target *gorm.DB, direction string, batchSize int, p events.Producer) *Validator[T] {
	res := &Validator[T]{
		base:      base,
		target:    target,
		direction: direction,
		batchSize: batchSize,
		p:         p,
	}
	res.fromBase = res.fullFromBase
	return res
}

func (v *Validator[T]) Validate(ctx context.Context) error {
	var eg errgroup.Group
	eg.Go(func() error {
		return v.validateBaseToTarget(ctx)
	})

	eg.Go(func() error {
		return v.validateTargetToBase(ctx)
	})

	return eg.Wait()
}

func (v *Validator[T]) validateBaseToTarget(ctx context.Context) error {
	offset := -1
	for {
		offset++
		src, err := v.fromBase(ctx, offset)
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 没有数据
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
			continue
		}
		if err != nil {
			// 查询出错了
			log.Println("base -> target 查询 src 失败", err)
			offset++
			continue
		}

		// 正常情况，进行校验
		var dst T
		err = v.target.WithContext(ctx).Where("id = ?", src.ID()).First(&dst).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			// target 没有数据, 向kafka丢一条消息
			v.notify(src.ID(), events.InconsistentEventTypeTargetMissing)

		case err == nil:
			// 正常情况，进行校验
			if !src.CompareTo(dst) {
				v.notify(src.ID(), events.InconsistentEventTypeNEQ)
			}
		default:
			// 其他情况, 直接记录日志
			log.Printf("base -> target 查询 dst 失败, id: %d, err: %v", src.ID(), err)
		}
	}
}

// 反向校验
func (v *Validator[T]) validateTargetToBase(ctx context.Context) error {
	offset := 0
	for {
		var ts []T
		err := v.target.WithContext(ctx).Select("id").
			Where("utime >?", v.utime).
			Order("id").
			Offset(offset).
			Limit(v.batchSize).
			Find(&ts).Error
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil
		}
		if errors.Is(err, gorm.ErrRecordNotFound) || len(ts) == 0 {
			// 没有数据
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
			continue
		}
		if err != nil {
			// 查询出错了
			log.Printf("target -> base 查询 tgt 失败, err: %v", err)
			offset += len(ts)
			continue
		}

		// 校验源表
		var srcs []T
		ids := slice.Map(ts, func(idx int, t T) int64 {
			return t.ID()
		})
		err = v.base.WithContext(ctx).Where("id IN ?", ids).Find(&srcs).Error
		if errors.Is(err, gorm.ErrRecordNotFound) || len(srcs) == 0 {
			// 没有数据
			// 向kafka丢一条消息
			v.notifyBaseMissing(ts)
			offset += len(ts)
			continue
		}
		if err != nil {
			// 查询出错了
			log.Printf("target -> base 查询 src 失败, err: %v", err)
			offset += len(ts)
			continue
		}
		diff := slice.DiffSetFunc(ts, srcs, func(src, dst T) bool {
			return src.ID() == dst.ID()
		})
		v.notifyBaseMissing(diff)
		if len(ts) < v.batchSize {
			// 没有数据了
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
		}
		offset += len(ts)
	}
}

func (v *Validator[T]) notify(id int64, typ string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := v.p.ProduceInconsistentEvent(ctx, events.InconsistentEvent{
		ID:        id,
		Direction: v.direction,
		Type:      typ,
	})
	if err != nil {
		log.Printf("通知消息不一致失败，id: %d, type: %s, err: %v\n", id, typ, err)
	}
}

func (v *Validator[T]) SleepInterval(sleetInterval time.Duration) *Validator[T] {
	v.sleepInterval = sleetInterval
	return v
}

func (v *Validator[T]) Utime(t int64) *Validator[T] {
	v.utime = t
	return v
}

func (v *Validator[T]) fullFromBase(ctx context.Context, offset int) (T, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	var src T
	err := v.base.WithContext(ctx).Order("id").Offset(offset).First(&src).Error
	return src, err
}

func (v *Validator[T]) incrFromBase(ctx context.Context, offset int) (T, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	var src T
	err := v.base.WithContext(ctx).Where("utime > ?", v.utime).Order("utime").Offset(offset).First(&src).Error
	return src, err
}

func (v *Validator[T]) Full() *Validator[T] {
	v.fromBase = v.fullFromBase
	return v
}

func (v *Validator[T]) Incr() *Validator[T] {
	v.fromBase = v.incrFromBase
	return v
}

func (v *Validator[T]) notifyBaseMissing(ts []T) {
	for _, t := range ts {
		v.notify(t.ID(), events.InconsistentEventTypeBaseMissing)
	}
}
