package dao

import (
	"context"
	"errors"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

const (
	PatternSrcOnly  = "src_only"
	PatternDstOnly  = "dst_only"
	PatternSrcFirst = "src_first"
	PatternDstFirst = "dst_first"
)

var errUnknownPattern = errors.New("未知的双写模式")

type DoubleWriteDao struct {
	src     InteractiveDAO
	dst     InteractiveDAO
	pattern *atomic.String
}

func (dao *DoubleWriteDao) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	pattern := dao.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		return dao.src.IncrReadCnt(ctx, biz, bizId)
	case PatternDstOnly:
		return dao.dst.IncrReadCnt(ctx, biz, bizId)
	case PatternSrcFirst:
		err := dao.src.IncrReadCnt(ctx, biz, bizId)
		if err != nil {
			return err
		}
		err = dao.dst.IncrReadCnt(ctx, biz, bizId)
		if err != nil {
			zap.L().Error("双写 read_cnt 写入dst失败", zap.Error(err), zap.String("biz", biz), zap.Int64("bizId", bizId))
		}
		return nil
	case PatternDstFirst:
		err := dao.dst.IncrReadCnt(ctx, biz, bizId)
		if err == nil {
			err1 := dao.src.IncrReadCnt(ctx, biz, bizId)
			if err1 != nil {
				zap.L().Error("双写 read_cnt 写入src失败", zap.Error(err1), zap.String("biz", biz), zap.Int64("bizId", bizId))
			}
		}
		return err
	default:
		return errUnknownPattern
	}
}

func (dao *DoubleWriteDao) InsertLikeInfo(ctx context.Context, biz string, id int64, uid int64) error {
	pattern := dao.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		return dao.src.InsertLikeInfo(ctx, biz, id, uid)
	case PatternDstOnly:
		return dao.dst.InsertLikeInfo(ctx, biz, id, uid)
	case PatternSrcFirst:
		err := dao.src.InsertLikeInfo(ctx, biz, id, uid)
		if err != nil {
			return err
		}
		err = dao.dst.InsertLikeInfo(ctx, biz, id, uid)
		if err != nil {
			zap.L().Error("双写 like_info 写入dst失败", zap.Error(err), zap.String("biz", biz), zap.Int64("uid", uid))
		}
		return nil
	case PatternDstFirst:
		err := dao.dst.InsertLikeInfo(ctx, biz, id, uid)
		if err == nil {
			err1 := dao.src.InsertLikeInfo(ctx, biz, id, uid)
			if err1 != nil {
				zap.L().Error("双写 like_info 写入src失败", zap.Error(err1), zap.String("biz", biz), zap.Int64("uid", uid))
			}
		}
		return err
	default:
		return errUnknownPattern
	}
}

func (dao *DoubleWriteDao) DeleteLikeInfo(ctx context.Context, biz string, id int64, uid int64) error {
	pattern := dao.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		return dao.src.DeleteLikeInfo(ctx, biz, id, uid)
	case PatternDstOnly:
		return dao.dst.DeleteLikeInfo(ctx, biz, id, uid)
	case PatternSrcFirst:
		err := dao.src.DeleteLikeInfo(ctx, biz, id, uid)
		if err != nil {
			return err
		}
		err = dao.dst.DeleteLikeInfo(ctx, biz, id, uid)
		if err != nil {
			zap.L().Error("双写 like_info 删除dst失败", zap.Error(err), zap.String("biz", biz), zap.Int64("uid", uid))
		}
		return nil
	case PatternDstFirst:
		err := dao.dst.DeleteLikeInfo(ctx, biz, id, uid)
		if err == nil {
			err1 := dao.src.DeleteLikeInfo(ctx, biz, id, uid)
			if err1 != nil {
				zap.L().Error("双写 like_info 删除src失败", zap.Error(err1), zap.String("biz", biz), zap.Int64("uid", uid))
			}
		}
		return err
	default:
		return errUnknownPattern
	}
}

func (dao *DoubleWriteDao) InsertCollectionBiz(ctx context.Context, biz string, bizId int64, cid int64, uid int64) error {
	pattern := dao.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		return dao.src.InsertCollectionBiz(ctx, biz, bizId, cid, uid)
	case PatternDstOnly:
		return dao.dst.InsertCollectionBiz(ctx, biz, bizId, cid, uid)
	case PatternSrcFirst:
		err := dao.src.InsertCollectionBiz(ctx, biz, bizId, cid, uid)
		if err != nil {
			return err
		}
		err = dao.dst.InsertCollectionBiz(ctx, biz, bizId, cid, uid)
		if err != nil {
			zap.L().Error("双写 collection_biz 写入dst失败", zap.Error(err), zap.String("biz", biz), zap.Int64("bizId", bizId), zap.Int64("cid", cid), zap.Int64("uid", uid))
		}
		return nil
	case PatternDstFirst:
		err := dao.dst.InsertCollectionBiz(ctx, biz, bizId, cid, uid)
		if err == nil {
			err1 := dao.src.InsertCollectionBiz(ctx, biz, bizId, cid, uid)
			if err1 != nil {
				zap.L().Error("双写 collection_biz 写入src失败", zap.Error(err1), zap.String("biz", biz), zap.Int64("bizId", bizId), zap.Int64("cid", cid), zap.Int64("uid", uid))
			}
		}
		return err
	default:
		return errUnknownPattern
	}
}

func (dao *DoubleWriteDao) Get(ctx context.Context, biz string, bizId int64) (Interactive, error) {
	pattern := dao.pattern.Load()
	switch pattern {
	case PatternSrcOnly, PatternSrcFirst:
		return dao.src.Get(ctx, biz, bizId)
	case PatternDstOnly, PatternDstFirst:
		return dao.dst.Get(ctx, biz, bizId)
	default:
		return Interactive{}, errUnknownPattern
	}
}

func (dao *DoubleWriteDao) GetLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) (UserLikeBiz, error) {
	pattern := dao.pattern.Load()
	switch pattern {
	case PatternSrcOnly, PatternSrcFirst:
		return dao.src.GetLikeInfo(ctx, biz, bizId, uid)
	case PatternDstOnly, PatternDstFirst:
		return dao.dst.GetLikeInfo(ctx, biz, bizId, uid)
	default:
		return UserLikeBiz{}, errUnknownPattern
	}
}

func (dao *DoubleWriteDao) GetCollectInfo(ctx context.Context, biz string, bizId int64, uid int64) (UserCollectionBiz, error) {
	pattern := dao.pattern.Load()
	switch pattern {
	case PatternSrcOnly, PatternSrcFirst:
		return dao.src.GetCollectInfo(ctx, biz, bizId, uid)
	case PatternDstOnly, PatternDstFirst:
		return dao.dst.GetCollectInfo(ctx, biz, bizId, uid)
	default:
		return UserCollectionBiz{}, errUnknownPattern
	}
}

func (dao *DoubleWriteDao) GetByIds(ctx context.Context, biz string, ids []int64) ([]Interactive, error) {
	pattern := dao.pattern.Load()
	switch pattern {
	case PatternSrcOnly, PatternSrcFirst:
		return dao.src.GetByIds(ctx, biz, ids)
	case PatternDstOnly, PatternDstFirst:
		return dao.dst.GetByIds(ctx, biz, ids)
	default:
		return nil, errUnknownPattern
	}
}

func NewDoubleWriteDao(src InteractiveDAO, dst InteractiveDAO) *DoubleWriteDao {
	return &DoubleWriteDao{
		src:     src,
		dst:     dst,
		pattern: atomic.NewString(PatternSrcOnly),
	}
}

func (dao *DoubleWriteDao) UpdatePattern(pattern string) {
	dao.pattern.Store(pattern)
}
