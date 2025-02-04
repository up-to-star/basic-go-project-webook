package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type JobDAO interface {
	Release(ctx context.Context, id int64) error
	UpdateTime(ctx context.Context, id int64) error
	UpdateNextTime(ctx context.Context, id int64, next time.Time) error
	Preempt(ctx context.Context) (Job, error)
}

type GORMJobDAO struct {
	db *gorm.DB
}

func (dao *GORMJobDAO) Release(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Model(&Job{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"utime":  now,
			"status": jobStatusWaiting,
		}).Error
}

func (dao *GORMJobDAO) UpdateTime(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Model(&Job{}).
		Where("id =?", id).
		Update("utime", now).Error
}

func (dao *GORMJobDAO) UpdateNextTime(ctx context.Context, id int64, next time.Time) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Model(&Job{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"utime":     now,
			"next_time": next.UnixMilli(),
		}).Error
}

func (dao *GORMJobDAO) Preempt(ctx context.Context) (Job, error) {
	db := dao.db.WithContext(ctx)
	for {
		var job Job
		now := time.Now().UnixMilli()
		err := db.Where("status = ? AND next_time < ?", jobStatusWaiting, now).First(&job).Error
		if err != nil {
			return job, err
		}
		res := db.Where("id = ? AND version = ?", job.Id, job.Version).
			Updates(map[string]interface{}{
				"status":  jobStatusRunning,
				"version": job.Version + 1,
				"utime":   now,
			})
		if res.Error != nil {
			return Job{}, res.Error
		}
		// 没有抢到
		if res.RowsAffected == 0 {
			continue
		}
		return job, nil
	}
}

func NewGORMJobDAO(db *gorm.DB) JobDAO {
	return &GORMJobDAO{
		db: db,
	}
}

type Job struct {
	Id         int64  `gorm:"primaryKey,autoIncrement"`
	Name       string `gorm:"type:varchar(128);unique"`
	Expression string
	Executor   string
	Cfg        string
	Status     int
	Ctime      int64
	Utime      int64
	NextTime   int64 `gorm:"index"`
	Version    int
}

const (
	jobStatusWaiting = iota
	jobStatusRunning
	jobStatusPaused
)
