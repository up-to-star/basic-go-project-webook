package article

import (
	"basic-project/webook/internal/domain"
	"basic-project/webook/internal/repository"
	"basic-project/webook/internal/repository/cache"
	"basic-project/webook/internal/repository/dao/article"
	"context"
	"go.uber.org/zap"
	"time"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	Sync(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx context.Context, id int64, authorId int64, status domain.ArticleStatus) error
	List(ctx context.Context, uid int64, limit int, offset int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPubById(ctx context.Context, id int64) (domain.Article, error)
}

type CachedArticleRepository struct {
	dao      article.ArticleDAO
	userRepo repository.UserRepository
	cache    cache.ArticleCache
}

func NewArticleRepository(dao article.ArticleDAO, cache cache.ArticleCache, userRepo repository.UserRepository) ArticleRepository {
	return &CachedArticleRepository{
		dao:      dao,
		cache:    cache,
		userRepo: userRepo,
	}
}

func (c *CachedArticleRepository) GetPubById(ctx context.Context, id int64) (domain.Article, error) {
	data, err := c.cache.GetPub(ctx, id)
	if err == nil {
		return data, err
	}
	art, err := c.dao.GetPubById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	res := pubToDomain(art)
	author, err := c.userRepo.FindById(ctx, art.AuthorId)
	if err != nil {
		zap.L().Warn("查询文章作者信息失败", zap.Error(err))
		return domain.Article{}, err
	}
	res.Author.Name = author.Nickname
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		er := c.cache.SetPub(ctx, res)
		if er != nil {
			zap.L().Warn("设置发布文章缓存失败", zap.Error(er))
		}
	}()
	return res, nil
}

func (c *CachedArticleRepository) GetById(ctx context.Context, id int64) (domain.Article, error) {
	res, err := c.cache.Get(ctx, id)
	if err == nil {
		return res, nil
	}
	art, err := c.dao.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	res = toDomain(art)
	go func() {
		err = c.cache.Set(ctx, res)
		if err != nil {
			zap.L().Error("根据文章ID设置缓存失败", zap.Int64("ID", id), zap.Error(err))
		}
	}()
	return res, nil
}

func (c *CachedArticleRepository) List(ctx context.Context, uid int64, limit int, offset int) ([]domain.Article, error) {
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

	// 缓存第一个文章
	go func() {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		c.preCache(ctx, res)
	}()

	return res, nil
}

func (c *CachedArticleRepository) SyncStatus(ctx context.Context, id int64, authorId int64, status domain.ArticleStatus) error {
	defer func() {
		err := c.cache.DeleteFirstPage(ctx, authorId)
		if err != nil {
			zap.L().Warn("删除文章list缓存失败", zap.Int64("art.id", authorId), zap.Error(err))
		}
		err = c.cache.Del(ctx, id)
		if err != nil {
			zap.L().Warn("删除缓存文章失败", zap.Int64("art.id", authorId), zap.Error(err))
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
			zap.L().Error("删除文章list缓存失败", zap.Int64("art.author_id", art.Author.Id), zap.Error(err))
		}
		err = c.cache.Del(ctx, art.Id)
		if err != nil {
			zap.L().Warn("删除文章缓存失败", zap.Int64("art.id", art.Id), zap.Error(err))
		}
	}()
	return c.dao.Sync(ctx, toArticleEntity(art))
}

func (c *CachedArticleRepository) Update(ctx context.Context, art domain.Article) error {
	defer func() {
		err := c.cache.DeleteFirstPage(ctx, art.Author.Id)
		if err != nil {
			zap.L().Error("删除文章list缓存失败", zap.Int64("art.author_id", art.Author.Id), zap.Error(err))
		}
		err = c.cache.Del(ctx, art.Id)
		if err != nil {
			zap.L().Warn("删除文章缓存失败", zap.Int64("art.id", art.Id), zap.Error(err))
		}
	}()
	return c.dao.UpdateById(ctx, toArticleEntity(art))
}

func (c *CachedArticleRepository) preCache(ctx context.Context, arts []domain.Article) {
	const size = 1024 * 1024
	if len(arts) > 0 && len(arts[0].Content) < size {
		err := c.cache.Set(ctx, arts[0])
		if err != nil {
			zap.L().Warn("缓存第一个文章失败", zap.Int64("author_id", arts[0].Author.Id), zap.Error(err))
		}
	}
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

func pubToDomain(art article.PublishedArticle) domain.Article {
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
