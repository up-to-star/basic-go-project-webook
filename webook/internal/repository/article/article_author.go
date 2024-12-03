package article

import (
	"basic-project/webook/internal/domain"
	"basic-project/webook/internal/repository/dao/article"
	"context"
)

type ArticleAuthorRepository interface {
	Update(ctx context.Context, art domain.Article) error
	Create(ctx context.Context, art domain.Article) (int64, error)
}

type CachedArticleAuthorRepository struct {
	dao article.ArticleAuthorDAO
}

func NewCachedArticleAuthorRepository(dao article.ArticleAuthorDAO) ArticleAuthorRepository {
	return &CachedArticleAuthorRepository{
		dao: dao,
	}
}

func (c *CachedArticleAuthorRepository) Update(ctx context.Context, art domain.Article) error {
	return c.dao.UpdateById(ctx, toArticleEntity(art))
}

func (c *CachedArticleAuthorRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	return c.dao.Insert(ctx, toArticleEntity(art))
}
