package repository

import (
	"basic-project/webook/interactive/domain"
	"basic-project/webook/interactive/repository/cache"
	"basic-project/webook/interactive/repository/dao"
	"context"
	"errors"
	"go.uber.org/zap"
)

type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	IncrLike(ctx context.Context, biz string, id int64, uid int64) error
	DecrLike(ctx context.Context, biz string, id int64, uid int64) error
	AddCollectionItem(ctx context.Context, biz string, bizId int64, cid int64, uid int64) error
	Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error)
	Liked(ctx context.Context, biz string, bizId int64, uid int64) (bool, error)
	Collected(ctx context.Context, biz string, bizId int64, uid int64) (bool, error)
	GetByIds(ctx context.Context, biz string, ids []int64) ([]domain.Interactive, error)
}

type CachedInteractiveRepository struct {
	dao   dao.InteractiveDAO
	cache cache.InteractiveCache
}

func (c *CachedInteractiveRepository) GetByIds(ctx context.Context, biz string, ids []int64) ([]domain.Interactive, error) {
	intrs, err := c.dao.GetByIds(ctx, biz, ids)
	if err != nil {
		return nil, err
	}

	res := make([]domain.Interactive, len(intrs))
	for i, intr := range intrs {
		res[i] = c.toDomain(intr)
	}
	return res, nil
}

func (c *CachedInteractiveRepository) Liked(ctx context.Context, biz string, bizId int64, uid int64) (bool, error) {
	_, err := c.dao.GetLikeInfo(ctx, biz, bizId, uid)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, dao.ErrRecordNotFount):
		return false, nil
	default:
		return false, err
	}
}

func (c *CachedInteractiveRepository) Collected(ctx context.Context, biz string, bizId int64, uid int64) (bool, error) {
	_, err := c.dao.GetCollectInfo(ctx, biz, bizId, uid)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, dao.ErrRecordNotFount):
		return false, nil
	default:
		return false, err
	}
}

func (c *CachedInteractiveRepository) Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
	intr, err := c.cache.Get(ctx, biz, bizId)
	if err == nil {
		return intr, nil
	}
	ie, err := c.dao.Get(ctx, biz, bizId)
	if err != nil {
		return domain.Interactive{}, err
	}

	res := c.toDomain(ie)
	err = c.cache.Set(ctx, biz, bizId, res)
	if err != nil {
		zap.L().Error("回写缓存失败", zap.Error(err), zap.Int64("bizId", bizId),
			zap.String("biz", biz))
	}
	return res, nil
}

func (c *CachedInteractiveRepository) AddCollectionItem(ctx context.Context, biz string, bizId int64, cid int64, uid int64) error {
	err := c.dao.InsertCollectionBiz(ctx, biz, bizId, cid, uid)
	if err != nil {
		return err
	}
	return c.cache.IncrCollectionCntIfPresent(ctx, biz, bizId)
}

func (c *CachedInteractiveRepository) DecrLike(ctx context.Context, biz string, id int64, uid int64) error {
	err := c.dao.DeleteLikeInfo(ctx, biz, id, uid)
	if err != nil {
		return err
	}
	return c.cache.DecrLikeCntIfPresent(ctx, biz, id)
}

func (c *CachedInteractiveRepository) IncrLike(ctx context.Context, biz string, id int64, uid int64) error {
	err := c.dao.InsertLikeInfo(ctx, biz, id, uid)
	if err != nil {
		return err
	}
	return c.cache.IncrLikeCntIfPresent(ctx, biz, id)
}

func (c *CachedInteractiveRepository) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	err := c.dao.IncrReadCnt(ctx, biz, bizId)
	if err != nil {
		return err
	}
	err = c.cache.IncrReadCntIfPresent(ctx, biz, bizId)
	return err
}

func (c *CachedInteractiveRepository) toDomain(interactive dao.Interactive) domain.Interactive {
	return domain.Interactive{
		ReadCnt:    interactive.ReadCnt,
		LikeCnt:    interactive.LikeCnt,
		CollectCnt: interactive.CollectCnt,
	}
}

func NewCachedInteractiveRepository(dao dao.InteractiveDAO, cache cache.InteractiveCache) InteractiveRepository {
	return &CachedInteractiveRepository{
		dao:   dao,
		cache: cache,
	}
}
