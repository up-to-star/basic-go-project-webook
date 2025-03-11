package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type GORMFollowDAO struct {
	db *gorm.DB
}

func (dao *GORMFollowDAO) CntFollower(ctx context.Context, uid int64) (int64, error) {
	var res int64
	err := dao.db.WithContext(ctx).
		Select("count(follower)").
		Where("follower = ? AND status = ?", uid, FollowRelationStatusActive).
		Count(&res).Error
	return res, err
}

func (dao *GORMFollowDAO) CntFollowee(ctx context.Context, uid int64) (int64, error) {
	var res int64
	err := dao.db.WithContext(ctx).
		Select("count(followee)").
		Where("followee = ? AND status = ?", uid, FollowRelationStatusActive).
		Count(&res).Error
	return res, err
}

func (dao *GORMFollowDAO) FollowRelationDetail(ctx context.Context, follower int64, followee int64) (FollowRelation, error) {
	var res FollowRelation
	err := dao.db.WithContext(ctx).
		Where("follower = ? AND followee = ? AND status = ?", follower, followee, FollowRelationStatusActive).
		First(&res).Error
	return res, err
}

func (dao *GORMFollowDAO) GetFollowee(ctx context.Context, follower int64, offset int64, limit int64) ([]FollowRelation, error) {
	var res []FollowRelation
	err := dao.db.WithContext(ctx).Where("follower = ? AND status = ?", follower, FollowRelationStatusActive).
		Offset(int(offset)).Limit(int(limit)).Find(&res).Error
	return res, err
}

func (dao *GORMFollowDAO) UpdateStatus(ctx context.Context, followee int64, follower int64, status uint8) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).
		Where("followee = ? AND follower = ?", followee, follower).
		Updates(map[string]interface{}{
			"status": status,
			"utime":  now,
		}).Error
}

func (dao *GORMFollowDAO) CreateFollowRelation(ctx context.Context, followee int64, follower int64) error {
	now := time.Now().UnixMilli()
	f := FollowRelation{
		Followee: followee,
		Follower: follower,
		Ctime:    now,
		Utime:    now,
	}
	return dao.db.WithContext(ctx).Clauses(clause.OnConflict{
		DoUpdates: clause.Assignments(map[string]interface{}{
			"status": FollowRelationStatusActive,
			"utime":  now,
		}),
	}).Create(&f).Error
}

func NewGORMFollowDAO(db *gorm.DB) FollowDAO {
	return &GORMFollowDAO{
		db: db,
	}
}
