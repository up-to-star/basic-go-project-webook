package ioc

import (
	"github.com/basic-go-project-webook/webook/internal/pkg/ratelimit"
	"github.com/redis/go-redis/v9"
	"time"
)

func InitLimiter(redisClient redis.Cmdable, interval time.Duration, rate int) ratelimit.Limiter {
	return ratelimit.NewRedisSlideWindowLimiter(redisClient, interval, rate)
}
