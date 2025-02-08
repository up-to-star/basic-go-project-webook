package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/basic-go-project-webook/webook/internal/domain"
	"github.com/redis/go-redis/v9"
	"time"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrKeyNotExists = errors.New("firstKey not exists")
)

type UserCache interface {
	Get(ctx context.Context, id int64) (domain.User, error)
	Set(ctx context.Context, user domain.User) error
	Del(ctx context.Context, id int64) error
}

type RedisUserCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

func NewUserCache(client redis.Cmdable) UserCache {
	return &RedisUserCache{
		client:     client,
		expiration: time.Minute * 15,
	}
}

func (cache *RedisUserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	key := cache.key(id)
	// 数据不存在 err = redis.Nil
	val, err := cache.client.Get(ctx, key).Bytes()
	if err != nil {
		return domain.User{}, ErrKeyNotExists
	}
	var user domain.User
	err = json.Unmarshal(val, &user)
	return user, err
}

func (cache *RedisUserCache) Set(ctx context.Context, user domain.User) error {
	val, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return cache.client.Set(ctx, cache.key(user.Id), val, cache.expiration).Err()
}

func (cache *RedisUserCache) Del(ctx context.Context, id int64) error {
	return cache.client.Del(ctx, cache.key(id)).Err()
}

func (cache *RedisUserCache) key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}
