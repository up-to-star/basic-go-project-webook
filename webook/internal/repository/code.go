package repository

import (
	"context"
	"github.com/basic-go-project-webook/webook/internal/repository/cache"
)

var (
	ErrCodeSendTooMany   = cache.ErrCodeSendTooMany
	ErrCodeVerifyTooMany = cache.ErrCodeVerifyTooMany
)

type CodeRepository interface {
	Store(ctx context.Context, biz, phong, code string) error
	Verify(ctx context.Context, biz, phong, expectedCode string) (bool, error)
}

type CachedCodeRepository struct {
	cache cache.CodeCache
}

func NewCodeRepository(cache cache.CodeCache) CodeRepository {
	return &CachedCodeRepository{
		cache: cache,
	}
}

func (repo *CachedCodeRepository) Store(ctx context.Context, biz, phong, code string) error {
	return repo.cache.Set(ctx, biz, phong, code)
}

func (repo *CachedCodeRepository) Verify(ctx context.Context, biz, phong, expectedCode string) (bool, error) {
	return repo.cache.Verify(ctx, biz, phong, expectedCode)
}
