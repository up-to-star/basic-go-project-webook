package repository

import (
	"basic-project/webook/internal/domain"
	"basic-project/webook/internal/repository/dao"
	"context"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
}

type CachedArticleRepository struct {
	dao dao.ArticleDAO
}

func NewArticleRepository(dao dao.ArticleDAO) ArticleRepository {
	return &CachedArticleRepository{
		dao: dao,
	}
}

func (c *CachedArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	return c.dao.Insert(ctx, dao.Article{
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
	})
}
