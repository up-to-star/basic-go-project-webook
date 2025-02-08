package cache

import (
	"basic-project/webook/interactive/domain"
	"context"
	_ "embed"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

var (
	//go:embed lua/incr_cnt.lua
	luaIncrCnt string
)

const (
	fieldReadCnt    = "read_cnt"
	fieldLikeCnt    = "like_cnt"
	fieldCollectCnt = "collect_cnt"
)

type InteractiveCache interface {
	IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error
	DecrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error
	IncrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error
	IncrCollectionCntIfPresent(ctx context.Context, biz string, bizId int64) error
	Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error)
	Set(ctx context.Context, biz string, bizId int64, inter domain.Interactive) error
}

type InteractiveRedisCache struct {
	client redis.Cmdable
}

func (c *InteractiveRedisCache) Set(ctx context.Context, biz string, bizId int64, inter domain.Interactive) error {
	key := c.key(biz, bizId)
	err := c.client.HSet(ctx, key, fieldCollectCnt, inter.CollectCnt,
		fieldReadCnt, inter.ReadCnt,
		fieldLikeCnt, inter.LikeCnt).Err()
	if err != nil {
		return err
	}
	return c.client.Expire(ctx, key, time.Minute*15).Err()
}

func (c *InteractiveRedisCache) Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
	key := c.key(biz, bizId)
	res, err := c.client.HGetAll(ctx, key).Result()
	if err != nil {
		return domain.Interactive{}, err
	}
	if len(res) == 0 {
		return domain.Interactive{}, ErrKeyNotExists
	}
	var inter domain.Interactive
	inter.CollectCnt, _ = strconv.ParseInt(res[fieldCollectCnt], 10, 64)
	inter.ReadCnt, _ = strconv.ParseInt(res[fieldReadCnt], 10, 64)
	inter.LikeCnt, _ = strconv.ParseInt(res[fieldLikeCnt], 10, 64)
	return inter, nil
}

func (c *InteractiveRedisCache) IncrCollectionCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	key := c.key(biz, bizId)
	return c.client.Eval(ctx, luaIncrCnt, []string{key}, fieldCollectCnt, 1).Err()
}

func (c *InteractiveRedisCache) DecrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	key := c.key(biz, bizId)
	return c.client.Eval(ctx, luaIncrCnt, []string{key}, fieldLikeCnt, -1).Err()
}

func (c *InteractiveRedisCache) IncrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	key := c.key(biz, bizId)
	return c.client.Eval(ctx, luaIncrCnt, []string{key}, fieldReadCnt, 1).Err()
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
