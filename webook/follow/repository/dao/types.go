package dao

import (
	"context"
)

type FollowDAO interface {
	CreateFollowRelation(ctx context.Context, followee int64, follower int64) error
	UpdateStatus(ctx context.Context, followee int64, follower int64, status uint8) error
	GetFollowee(ctx context.Context, follower int64, offset int64, limit int64) ([]FollowRelation, error)
	FollowRelationDetail(ctx context.Context, follower int64, followee int64) (FollowRelation, error)
	CntFollower(ctx context.Context, uid int64) (int64, error)
	CntFollowee(ctx context.Context, uid int64) (int64, error)
}

const (
	FollowRelationStatusUnknown uint8 = iota
	FollowRelationStatusActive
	FollowRelationStatusInactive
)

type FollowRelation struct {
	Id       int64 `gorm:"column:id;autoIncrement;primaryKey;"`
	Follower int64 `gorm:"uniqueIndex:follower_followee"`
	Followee int64 `gorm:"uniqueIndex:follower_followee"`
	Status   uint8
	Ctime    int64
	Utime    int64
}

type FollowStatics struct {
	Id        int64 `gorm:"primaryKey,autoIncrement,column:id"`
	Uid       int64 `gorm:"unique"`
	Followees int64
	Followers int64
	Status    uint8
	Utime     int64
	Ctime     int64
}
