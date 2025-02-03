package service

import (
	"basic-project/webook/internal/domain"
	events "basic-project/webook/internal/events/article"
	"basic-project/webook/internal/repository/article"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"time"
)

type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	PublishV1(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx *gin.Context, art domain.Article) error
	List(ctx *gin.Context, uid int64, limit int, offset int) ([]domain.Article, error)
	ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]domain.Article, error)
	GetById(ctx *gin.Context, id int64) (domain.Article, error)
	GetPubById(ctx context.Context, id, uid int64) (domain.Article, error)
}

type articleService struct {
	repo     article.ArticleRepository
	producer events.Producer

	// v1
	authorRepo article.ArticleAuthorRepository
	readerRepo article.ArticleReaderRepository
}

func NewArticleService(repo article.ArticleRepository, producer events.Producer) ArticleService {
	return &articleService{
		repo:     repo,
		producer: producer,
	}
}

func (a *articleService) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]domain.Article, error) {
	return a.repo.ListPub(ctx, start, offset, limit)
}

func (a *articleService) GetPubById(ctx context.Context, id, uid int64) (domain.Article, error) {
	art, err := a.repo.GetPubById(ctx, id)
	if err == nil {
		go func() {
			a.producer.ProduceReadEvent(ctx, events.ReadEvent{
				Uid: uid,
				Aid: id,
			})
		}()
	}
	return art, err
}

func (a *articleService) GetById(ctx *gin.Context, id int64) (domain.Article, error) {
	return a.repo.GetById(ctx, id)
}

func (a *articleService) List(ctx *gin.Context, uid int64, limit int, offset int) ([]domain.Article, error) {
	return a.repo.List(ctx, uid, limit, offset)
}

func (a *articleService) Withdraw(ctx *gin.Context, art domain.Article) error {
	return a.repo.SyncStatus(ctx, art.Id, art.Author.Id, domain.ArticleStatusPrivate)
}
func (a *articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusPublished
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
	art.Status = domain.ArticleStatusUnpublished
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
