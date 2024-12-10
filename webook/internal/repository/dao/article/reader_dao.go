package article

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"time"
)

type ArticleReaderDAO interface {
	UpdateById(ctx context.Context, art PublishedArticle) error
	Insert(ctx context.Context, art PublishedArticle) error
}

type GORMArticleReaderDAO struct {
	db *gorm.DB
}

func NewGORMArticleReaderDAO(db *gorm.DB) *GORMArticleReaderDAO {
	return &GORMArticleReaderDAO{
		db: db,
	}
}

func (dao *GORMArticleReaderDAO) UpdateById(ctx context.Context, art PublishedArticle) error {
	now := time.Now().UnixMilli()
	art.Utime = now
	res := dao.db.WithContext(ctx).Model(&art).
		Where("id = ?", art.Id).Updates(map[string]any{
		"title":   art.Title,
		"content": art.Content,
		"utime":   art.Utime,
	})
	if res.RowsAffected == 0 {
		return fmt.Errorf("更新失败, id: %d", art.Id)
	}
	return res.Error
}

func (dao *GORMArticleReaderDAO) Insert(ctx context.Context, art PublishedArticle) error {
	now := time.Now().UnixMilli()
	art.Utime = now
	art.Ctime = now
	return dao.db.WithContext(ctx).Create(&art).Error
}
