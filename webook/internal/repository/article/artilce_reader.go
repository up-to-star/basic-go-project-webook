package article

import (
	"basic-project/webook/internal/domain"
	"basic-project/webook/internal/repository/dao/article"
	"context"
)

type ArticleReaderRepository interface {
	Save(ctx context.Context, art domain.Article) error
}

type CachedArticleReaderRepository struct {
	dao article.ArticleReaderDAO
}

func NewArticleReaderRepository(dao article.ArticleReaderDAO) *CachedArticleReaderRepository {
	return &CachedArticleReaderRepository{
		dao: dao,
	}
}

func (c *CachedArticleReaderRepository) Save(ctx context.Context, art domain.Article) error {
	if art.Id > 0 {
		return c.dao.UpdateById(ctx, toPublishedArticle(art))
	}
	return c.dao.Insert(ctx, toPublishedArticle(art))
}
