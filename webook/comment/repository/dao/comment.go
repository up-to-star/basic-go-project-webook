package dao

import (
	"context"
	"database/sql"
	"gorm.io/gorm"
)

type CommentDAO interface {
	Insert(ctx context.Context, comment Comment) error
	Delete(ctx context.Context, comment Comment) error
	FindByBiz(ctx context.Context, biz string, bizId int64, limit int64, minId int64) ([]Comment, error)
	FindRepliesByPid(ctx context.Context, pid int64, offset, limit int) ([]Comment, error)
	FindRepliesByRid(ctx context.Context, rid int64, limit int64, maxId int64) ([]Comment, error)
	GetCommentByIds(ctx context.Context, ids []int64) ([]Comment, error)
}

type GORMCommentDao struct {
	db *gorm.DB
}

func (dao *GORMCommentDao) GetCommentByIds(ctx context.Context, ids []int64) ([]Comment, error) {
	var res []Comment
	err := dao.db.WithContext(ctx).Where("id IN ?", ids).Find(&res).Error
	return res, err
}

func (dao *GORMCommentDao) FindRepliesByRid(ctx context.Context, rid int64, limit int64, maxId int64) ([]Comment, error) {
	var res []Comment
	err := dao.db.WithContext(ctx).
		Where("root_id = ? AND id > ?", rid, maxId).
		Order("id ASC").
		Limit(int(limit)).
		Find(&res).Error
	return res, err
}

func (dao *GORMCommentDao) FindRepliesByPid(ctx context.Context, pid int64, offset, limit int) ([]Comment, error) {
	var res []Comment
	err := dao.db.WithContext(ctx).Where("pid = ?", pid).
		Order("id DESC").
		Offset(offset).Limit(limit).Find(&res).Error
	return res, err
}

func (dao *GORMCommentDao) FindByBiz(ctx context.Context, biz string, bizId int64, limit int64, minId int64) ([]Comment, error) {
	var res []Comment
	err := dao.db.WithContext(ctx).
		Where("biz = ? AND biz_id = ? AND id < ? AND pid IS NULL", biz, bizId, minId).
		Limit(int(limit)).Find(&res).Error
	return res, err
}

func (dao *GORMCommentDao) Delete(ctx context.Context, comment Comment) error {
	return dao.db.WithContext(ctx).Delete(&Comment{
		Id: comment.Id,
	}).Error
}

func (dao *GORMCommentDao) Insert(ctx context.Context, comment Comment) error {
	return dao.db.WithContext(ctx).Create(&comment).Error
}

func NewCommentDAO(db *gorm.DB) CommentDAO {
	return &GORMCommentDao{
		db: db,
	}
}

type Comment struct {
	Id int64 `gorm:"autoIncrement,primaryKey"`
	// 发表评论的人，可以根据这个找到他所有的评论
	Uid     int64
	Biz     string `gorm:"index:biz_type_id"`
	BizId   int64  `gorm:"index:biz_type_id"`
	Content string

	// 我的根评论是哪个
	// 也就是说，如果这个字段是 NULL，它是根评论
	RootId sql.NullInt64 `gorm:"column:root_id;index"`
	PID    sql.NullInt64 `gorm:"column:pid;index"`

	// 关联删除
	ParentComment *Comment `gorm:"ForeignKey:PID;AssociationForeignKey:ID;constraint:OnDelete:CASCADE"`
	Ctime         int64
	Utime         int64
}
