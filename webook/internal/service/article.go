package service

import (
	"basic-project/webook/internal/domain"
	"basic-project/webook/internal/repository/article"
	"context"
	"errors"
	"go.uber.org/zap"
)

type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	PublishV1(ctx context.Context, art domain.Article) (int64, error)
}

type articleService struct {
	repo article.ArticleRepository

	// v1
	authorRepo article.ArticleAuthorRepository
	readerRepo article.ArticleReaderRepository
}

func NewArticleService(repo article.ArticleRepository) ArticleService {
	return &articleService{
		repo: repo,
	}
}

func (a *articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	return a.repo.Sync(ctx, art)
}

func (a *articleService) PublishV1(ctx context.Context, art domain.Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)
	if art.Id > 0 {
		err = a.authorRepo.Update(ctx, art)
	} else {
		id, err = a.authorRepo.Create(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	art.Id = id
	for i := 0; i < 3; i++ {
		err = a.readerRepo.Save(ctx, art)
		if err != nil {
			zap.L().Error("保存打制作库成功，但保存到线上库失败", zap.Error(err))
		} else {
			return id, nil
		}
	}
	zap.L().Error("保存到制作库成功，但保存到线上库重试全部失败", zap.Error(err))
	return id, errors.New("保存到线上库失败，重试次数用完")
}

func (a *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	if art.Id > 0 {
		err := a.repo.Update(ctx, art)
		return art.Id, err
	}
	return a.repo.Create(ctx, art)
}

func NewArticleServiceV1(repo article.ArticleRepository,
	readerRepo article.ArticleReaderRepository,
	authorRepo article.ArticleAuthorRepository) ArticleService {
	return &articleService{
		repo:       repo,
		readerRepo: readerRepo,
		authorRepo: authorRepo,
	}
}
