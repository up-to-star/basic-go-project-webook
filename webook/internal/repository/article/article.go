package article

import (
	"basic-project/webook/internal/domain"
	"basic-project/webook/internal/repository/cache"
	"basic-project/webook/internal/repository/dao/article"
	"context"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"time"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	Sync(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx *gin.Context, id int64, authorId int64, status domain.ArticleStatus) error
	List(ctx *gin.Context, uid int64, limit int, offset int) ([]domain.Article, error)
}

type CachedArticleRepository struct {
	dao   article.ArticleDAO
	cache cache.ArticleCache
}

func NewArticleRepository(dao article.ArticleDAO, cache cache.ArticleCache) ArticleRepository {
	return &CachedArticleRepository{
		dao:   dao,
		cache: cache,
	}
}

func (c *CachedArticleRepository) List(ctx *gin.Context, uid int64, limit int, offset int) ([]domain.Article, error) {
	if offset == 0 && limit <= 100 {
		res, err := c.cache.GetFirstPage(ctx, uid)
		if err == nil {
			return res, err
		} else {
			zap.L().Info("缓存未命中，没找到文章信息", zap.Int64("uid", uid), zap.Error(err))
		}
	}
	arts, err := c.dao.GetByAuthor(ctx, uid, limit, offset)
	if err != nil {
		return nil, err
	}
	res := make([]domain.Article, len(arts))
	for i, art := range arts {
		res[i] = toDomain(art)
	}

	go func() {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		if offset == 0 && limit <= 100 {
			err = c.cache.SetFirstPage(ctx, uid, res)
			if err != nil {
				zap.L().Error("文章写入缓存失败", zap.Int64("uid", uid), zap.Error(err))
			}
		}
	}()
	return res, nil
}

func (c *CachedArticleRepository) SyncStatus(ctx *gin.Context, id int64, authorId int64, status domain.ArticleStatus) error {
	defer func() {
		err := c.cache.DeleteFirstPage(ctx, authorId)
		if err != nil {
			zap.L().Error("删除文章缓存失败", zap.Int64("art.id", authorId), zap.Error(err))
		}
	}()
	return c.dao.SyncStatus(ctx, id, authorId, status.ToUint8())
}

func (c *CachedArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	defer func() {
		err := c.cache.DeleteFirstPage(ctx, art.Author.Id)
		if err != nil {
			zap.L().Error("删除文章缓存失败", zap.Int64("art.id", art.Author.Id), zap.Error(err))
		}
	}()
	return c.dao.Insert(ctx, toArticleEntity(art))
}

func (c *CachedArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	defer func() {
		err := c.cache.DeleteFirstPage(ctx, art.Author.Id)
		if err != nil {
			zap.L().Error("删除文章缓存失败", zap.Int64("art.id", art.Author.Id), zap.Error(err))
		}
	}()
	return c.dao.Sync(ctx, toArticleEntity(art))
}

func (c *CachedArticleRepository) Update(ctx context.Context, art domain.Article) error {
	defer func() {
		err := c.cache.DeleteFirstPage(ctx, art.Author.Id)
		if err != nil {
			zap.L().Error("删除文章缓存失败", zap.Int64("art.id", art.Author.Id), zap.Error(err))
		}
	}()
	return c.dao.UpdateById(ctx, toArticleEntity(art))
}

func toArticleEntity(art domain.Article) article.Article {
	return article.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	}
}

func toDomain(art article.Article) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Author: domain.Author{
			Id: art.AuthorId,
		},
		Ctime:  time.UnixMilli(art.Ctime),
		Utime:  time.UnixMilli(art.Utime),
		Status: domain.ArticleStatus(art.Status),
	}
}

func toPublishedArticle(art domain.Article) article.PublishedArticle {
	return article.PublishedArticle{
		Article: toArticleEntity(art),
	}
}
