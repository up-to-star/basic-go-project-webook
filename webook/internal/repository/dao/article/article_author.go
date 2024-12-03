package article

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"time"
)

type ArticleAuthorDAO interface {
	UpdateById(ctx context.Context, art Article) error
	Insert(ctx context.Context, art Article) (int64, error)
}

type GORMArticleAuthor struct {
	db *gorm.DB
}

func NewGORMArticleAuthor(db *gorm.DB) *GORMArticleAuthor {
	return &GORMArticleAuthor{
		db: db,
	}
}

func (dao *GORMArticleAuthor) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	art.Utime = now
	res := dao.db.WithContext(ctx).Model(&art).
		Where("id = ? AND author_id = ?", art.Id, art.AuthorId).Updates(map[string]any{
		"title":   art.Title,
		"content": art.Content,
		"utime":   art.Utime,
	})
	if res.RowsAffected == 0 {
		return fmt.Errorf("更新失败, 可能创作者非法, id: %d, author_id: %d", art.Id, art.AuthorId)
	}
	return res.Error
}

func (dao *GORMArticleAuthor) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := dao.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}
