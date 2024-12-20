package service

import (
	"basic-project/webook/internal/domain"
	"basic-project/webook/internal/repository"
	"context"
	"golang.org/x/sync/errgroup"
)

type InteractiveService interface {
	Like(ctx context.Context, biz string, id int64, uid int64) error
	CancelLike(ctx context.Context, biz string, id int64, uid int64) error
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	Collect(ctx context.Context, biz string, bizId int64, cid int64, uid int64) error
	Get(ctx context.Context, biz string, bizId int64, uid int64) (domain.Interactive, error)
}

type interactiveService struct {
	repo repository.InteractiveRepository
}

func (i *interactiveService) Get(ctx context.Context, biz string, bizId int64, uid int64) (domain.Interactive, error) {
	intr, err := i.repo.Get(ctx, biz, bizId)
	if err != nil {
		return domain.Interactive{}, err
	}

	var eg errgroup.Group
	eg.Go(func() error {
		liked, er := i.repo.Liked(ctx, biz, bizId, uid)
		intr.Liked = liked
		return er
	})

	eg.Go(func() error {
		collected, er := i.repo.Collected(ctx, biz, bizId, uid)
		intr.Collected = collected
		return er
	})
	return intr, eg.Wait()
}

func (i *interactiveService) Collect(ctx context.Context, biz string, bizId int64, cid int64, uid int64) error {
	return i.repo.AddCollectionItem(ctx, biz, bizId, cid, uid)
}

func (i *interactiveService) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	return i.repo.IncrReadCnt(ctx, biz, bizId)
}

func (i *interactiveService) CancelLike(ctx context.Context, biz string, id int64, uid int64) error {
	return i.repo.DecrLike(ctx, biz, id, uid)
}

func (i *interactiveService) Like(ctx context.Context, biz string, id int64, uid int64) error {
	return i.repo.IncrLike(ctx, biz, id, uid)
}

func NewInteractiveService(repo repository.InteractiveRepository) InteractiveService {
	return &interactiveService{
		repo: repo,
	}
}
