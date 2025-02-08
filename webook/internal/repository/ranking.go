package repository

import (
	"context"
	"github.com/basic-go-project-webook/webook/internal/domain"
	"github.com/basic-go-project-webook/webook/internal/repository/cache"
)

type RankingRepository interface {
	ReplaceTopN(ctx context.Context, arts []domain.Article) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
}

type OnlyCachedRankingRepository struct {
	redisCache *cache.RankingRedisCache
	localCache *cache.RankingLocalCache
}

func (c *OnlyCachedRankingRepository) GetTopN(ctx context.Context) ([]domain.Article, error) {
	data, err := c.localCache.Get(ctx)
	if err == nil {
		return data, err
	}
	data, err = c.redisCache.Get(ctx)
	if err != nil {
		return c.localCache.Get(ctx)
	}
	_ = c.localCache.Set(ctx, data)
	return data, nil
}

func (c *OnlyCachedRankingRepository) ReplaceTopN(ctx context.Context, arts []domain.Article) error {
	_ = c.localCache.Set(ctx, arts)
	return c.redisCache.Set(ctx, arts)
}

func NewOnlyCachedRankingRepository(redis *cache.RankingRedisCache, local *cache.RankingLocalCache) RankingRepository {
	return &OnlyCachedRankingRepository{
		redisCache: redis,
		localCache: local,
	}
}
