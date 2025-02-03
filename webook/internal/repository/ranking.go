package repository

import (
	"basic-project/webook/internal/domain"
	"basic-project/webook/internal/repository/cache"
	"context"
)

type RankingRepository interface {
	ReplaceTopN(ctx context.Context, arts []domain.Article) error
}

type OnlyCachedRankingRepository struct {
	cache cache.RankingCache
}

func (c *OnlyCachedRankingRepository) ReplaceTopN(ctx context.Context, arts []domain.Article) error {
	return c.cache.Set(ctx, arts)
}

func NewOnlyCachedRankingRepository(cache cache.RankingCache) RankingRepository {
	return &OnlyCachedRankingRepository{
		cache: cache,
	}
}
