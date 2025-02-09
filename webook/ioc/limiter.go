package ioc

import (
	ratelimit2 "github.com/basic-go-project-webook/webook/pkg/ratelimit"
	"github.com/redis/go-redis/v9"
	"time"
)

func InitLimiter(redisClient redis.Cmdable, interval time.Duration, rate int) ratelimit2.Limiter {
	return ratelimit2.NewRedisSlideWindowLimiter(redisClient, interval, rate)
}
