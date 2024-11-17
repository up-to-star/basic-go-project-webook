package repository

import (
	"basic-project/webook/internal/repository/cache"
	"context"
)

var (
	ErrCodeSendTooMany   = cache.ErrCodeSendTooMany
	ErrCodeVerifyTooMany = cache.ErrCodeVerifyTooMany
)

type CodeRepository struct {
	cache *cache.CodeCache
}

func NewCodeRepository(cache *cache.CodeCache) *CodeRepository {
	return &CodeRepository{
		cache: cache,
	}
}

func (repo *CodeRepository) Store(ctx context.Context, biz, phong, code string) error {
	return repo.cache.Set(ctx, biz, phong, code)
}

func (repo *CodeRepository) Verify(ctx context.Context, biz, phong, expectedCode string) (bool, error) {
	return repo.cache.Verify(ctx, biz, phong, expectedCode)
}
