package repository

import (
	"context"
	"github.com/basic-go-project-webook/webook/follow/domain"
	"github.com/basic-go-project-webook/webook/follow/repository/cache"
	"github.com/basic-go-project-webook/webook/follow/repository/dao"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type FollowRepository interface {
	AddFollowRelation(ctx context.Context, followee int64, follower int64) error
	InactiveFollowRelation(ctx context.Context, followee int64, follower int64) error
	GetFollowee(ctx context.Context, follower int64, offset int64, limit int64) ([]domain.FollowRelation, error)
	FollowInfo(ctx context.Context, follower int64, followee int64) (domain.FollowRelation, error)
	GetFollowStatics(ctx context.Context, uid int64) (domain.FollowStatics, error)
}

type CachedFollowRepository struct {
	dao   dao.FollowDAO
	cache cache.FollowCache
}

func (c *CachedFollowRepository) GetFollowStatics(ctx context.Context, uid int64) (domain.FollowStatics, error) {
	// 加入 redis 缓存
	res, err := c.cache.StaticsInfo(ctx, uid)
	if err == nil {
		return res, nil
	}
	var eg errgroup.Group
	eg.Go(func() error {
		followees, er := c.dao.CntFollowee(ctx, uid)
		if er != nil {
			return er
		}
		res.Followees = followees
		return nil
	})
	eg.Go(func() error {
		followers, er := c.dao.CntFollower(ctx, uid)
		if er != nil {
			return er
		}
		res.Followers = followers
		return nil
	})
	err = eg.Wait()
	if err != nil {
		return domain.FollowStatics{}, err
	}
	err = c.cache.SetStaticsInfo(ctx, uid, res)
	if err != nil {
		zap.L().Error("redis 写入失败", zap.Error(err))
	}
	return res, nil
}

func (c *CachedFollowRepository) FollowInfo(ctx context.Context, follower int64, followee int64) (domain.FollowRelation, error) {
	val, err := c.dao.FollowRelationDetail(ctx, follower, followee)
	if err != nil {
		return domain.FollowRelation{}, err
	}
	return c.toDomain(val), nil
}

func (c *CachedFollowRepository) GetFollowee(ctx context.Context, follower int64, offset int64, limit int64) ([]domain.FollowRelation, error) {
	followRelations, err := c.dao.GetFollowee(ctx, follower, offset, limit)
	if err != nil {
		return nil, err
	}
	res := make([]domain.FollowRelation, 0, len(followRelations))
	for _, followRelation := range followRelations {
		res = append(res, c.toDomain(followRelation))
	}
	return res, nil
}

func (c *CachedFollowRepository) InactiveFollowRelation(ctx context.Context, followee int64, follower int64) error {
	return c.dao.UpdateStatus(ctx, followee, follower, dao.FollowRelationStatusInactive)
}

func (c *CachedFollowRepository) AddFollowRelation(ctx context.Context, followee int64, follower int64) error {
	return c.dao.CreateFollowRelation(ctx, followee, follower)
}

func NewFollowRepository(dao dao.FollowDAO, cache cache.FollowCache) FollowRepository {
	return &CachedFollowRepository{
		dao:   dao,
		cache: cache,
	}
}

func (c *CachedFollowRepository) toDomain(followRelation dao.FollowRelation) domain.FollowRelation {
	return domain.FollowRelation{
		Followee: followRelation.Followee,
		Follower: followRelation.Follower,
	}
}
