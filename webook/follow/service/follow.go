package service

import (
	"context"
	"github.com/basic-go-project-webook/webook/follow/domain"
	"github.com/basic-go-project-webook/webook/follow/repository"
)

type FollowService interface {
	Follow(ctx context.Context, followee, follower int64) error
	CancelFollow(ctx context.Context, followee, follower int64) error
	GetFollowee(ctx context.Context, follower int64, offset int64, limit int64) ([]domain.FollowRelation, error)
	FollowInfo(ctx context.Context, follower int64, followee int64) (domain.FollowRelation, error)
	GetFollowStatics(ctx context.Context, uid int64) (domain.FollowStatics, error)
}
type followService struct {
	repo repository.FollowRepository
}

func (f *followService) GetFollowStatics(ctx context.Context, uid int64) (domain.FollowStatics, error) {
	return f.repo.GetFollowStatics(ctx, uid)
}

func (f *followService) FollowInfo(ctx context.Context, follower int64, followee int64) (domain.FollowRelation, error) {
	res, err := f.repo.FollowInfo(ctx, follower, followee)
	return res, err
}

func (f *followService) GetFollowee(ctx context.Context, follower int64, offset int64, limit int64) ([]domain.FollowRelation, error) {
	return f.repo.GetFollowee(ctx, follower, offset, limit)
}

func (f *followService) Follow(ctx context.Context, followee, follower int64) error {
	return f.repo.AddFollowRelation(ctx, followee, follower)
}

func (f *followService) CancelFollow(ctx context.Context, followee, follower int64) error {
	return f.repo.InactiveFollowRelation(ctx, followee, follower)
}

func NewFollowService(repo repository.FollowRepository) FollowService {
	return &followService{
		repo: repo,
	}
}
