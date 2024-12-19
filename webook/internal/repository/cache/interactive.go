package cache

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/redis/go-redis/v9"
)

var (
	//go:embed lua/incr_cnt.lua
	luaIncrCnt string
)

const (
	fieldReadCnt = "read_cnt"
)

type InteractiveCache interface {
	IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error
}

type InteractiveRedisCache struct {
	client redis.Cmdable
}

func (c *InteractiveRedisCache) IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	key := c.key(biz, bizId)
	return c.client.Eval(ctx, luaIncrCnt, []string{key}, fieldReadCnt, 1).Err()
}

func NewInteractiveRedisCache(client redis.Cmdable) InteractiveCache {
	return &InteractiveRedisCache{
		client: client,
	}
}

func (c *InteractiveRedisCache) key(biz string, bizId int64) string {
	return fmt.Sprintf("interactive:%s:%d", biz, bizId)
}
