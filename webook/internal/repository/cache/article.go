package cache

import (
	"basic-project/webook/internal/domain"
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"time"
)

type ArticleCache interface {
	GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error)
	SetFirstPage(ctx context.Context, uid int64, arts []domain.Article) error
	DeleteFirstPage(ctx context.Context, uid int64) error
}

type RedisArticleCache struct {
	client redis.Cmdable
}

func (r *RedisArticleCache) GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error) {
	data, err := r.client.Get(ctx, r.key(uid)).Bytes()
	if err != nil {
		zap.L().Error("缓存获取文章失败", zap.Int64("uid", uid), zap.Error(err))
		return nil, err
	}
	var arts []domain.Article
	err = json.Unmarshal(data, &arts)
	return arts, err
}

func (r *RedisArticleCache) SetFirstPage(ctx context.Context, uid int64, arts []domain.Article) error {
	for i := 0; i < len(arts); i++ {
		arts[i].Content = arts[i].Abstract()
	}

	data, err := json.Marshal(arts)
	if err != nil {
		zap.L().Error("arts json marshal failed", zap.Error(err))
		return err
	}

	return r.client.Set(ctx, r.key(uid), data, time.Minute*10).Err()
}

func (r *RedisArticleCache) DeleteFirstPage(ctx context.Context, uid int64) error {
	return r.client.Del(ctx, r.key(uid)).Err()
}

func (r *RedisArticleCache) key(uid int64) string {
	return fmt.Sprintf("article:first_page:%d", uid)
}

func NewRedisArticleCache(client redis.Cmdable) ArticleCache {
	return &RedisArticleCache{
		client: client,
	}
}
