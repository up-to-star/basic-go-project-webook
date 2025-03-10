package article

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type ArticleDAO interface {
	Insert(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, art Article) error
	Sync(ctx context.Context, art Article) (int64, error)
	SyncStatus(ctx context.Context, id int64, authorId int64, status uint8) error
	GetByAuthor(ctx context.Context, uid int64, limit int, offset int) ([]Article, error)
	GetById(ctx context.Context, id int64) (Article, error)
	GetPubById(ctx context.Context, id int64) (PublishedArticle, error)
	ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]Article, error)
}

type GORMArticleDAO struct {
	db *gorm.DB
}

func NewArticleDAO(db *gorm.DB) ArticleDAO {
	return &GORMArticleDAO{
		db: db,
	}
}

func (dao *GORMArticleDAO) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]Article, error) {
	var res []Article
	err := dao.db.WithContext(ctx).
		Where("utime < ?", start.UnixMilli()).
		Order("utime DESC").
		Offset(offset).
		Limit(limit).
		Find(&res).Error
	return res, err
}

func (dao *GORMArticleDAO) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	var art PublishedArticle
	err := dao.db.WithContext(ctx).Where("id = ?", id).First(&art).Error
	if err != nil {
		return PublishedArticle{}, err
	}
	return art, nil
}

func (dao *GORMArticleDAO) GetById(ctx context.Context, id int64) (Article, error) {
	var art Article
	err := dao.db.WithContext(ctx).Where("id = ?", id).First(&art).Error
	if err != nil {
		return Article{}, err
	}
	return art, nil
}

func (dao *GORMArticleDAO) GetByAuthor(ctx context.Context, uid int64, limit int, offset int) ([]Article, error) {
	var arts []Article
	err := dao.db.WithContext(ctx).Model(&Article{}).
		Where("author_id = ?", uid).
		Offset(offset).
		Limit(limit).
		Order("utime DESC").
		Find(&arts).Error
	return arts, err
}

func (dao *GORMArticleDAO) SyncStatus(ctx context.Context, id int64, authorId int64, status uint8) error {
	now := time.Now().UnixMilli()
	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).Where("id = ? AND author_id = ?", id, authorId).
			Updates(map[string]interface{}{
				"status": status,
				"utime":  now,
			})
		if res.Error != nil {
			zap.L().Error("数据库错误", zap.Error(res.Error))
			return res.Error
		}
		if res.RowsAffected != 1 {
			zap.L().Error("ID 或者 authorId 错误")
			return fmt.Errorf("id 或者 authorId 错误, uid: %d, authorId: %d", id, authorId)
		}

		return tx.Model(&PublishedArticle{}).Where("id = ?", id).
			Updates(map[string]interface{}{
				"status": status,
				"utime":  now,
			}).Error
	})

	return err
}

func (dao *GORMArticleDAO) Sync(ctx context.Context, art Article) (int64, error) {
	var id = art.Id
	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		daoTX := NewArticleDAO(tx)
		if id > 0 {
			err = daoTX.UpdateById(ctx, art)
		} else {
			id, err = dao.Insert(ctx, art)
		}
		if err != nil {
			return err
		}
		art.Id = id
		now := time.Now().UnixMilli()
		pubArt := PublishedArticle{art}
		pubArt.Utime = now
		pubArt.Ctime = now

		err = tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"title":   pubArt.Title,
				"content": pubArt.Content,
				"utime":   pubArt.Utime,
				"status":  pubArt.Status,
			}),
		}).Create(&pubArt).Error
		return err
	})
	return id, err
}

func (dao *GORMArticleDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := dao.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}

func (dao *GORMArticleDAO) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	art.Utime = now
	res := dao.db.WithContext(ctx).Model(&art).
		Where("id = ? AND author_id = ?", art.Id, art.AuthorId).Updates(map[string]any{
		"title":   art.Title,
		"content": art.Content,
		"utime":   art.Utime,
		"status":  art.Status,
	})
	if res.RowsAffected == 0 {
		return fmt.Errorf("更新失败, 可能创作者非法, id: %d, author_id: %d", art.Id, art.AuthorId)
	}
	return res.Error
}

// Article 制作库
type Article struct {
	Id       int64  `gorm:"primaryKey;autoIncrement" bson:"id,omitempty"`
	Title    string `gorm:"type:varchar(1024)" bson:"title,omitempty"`
	Content  string `gorm:"type:BLOB" bson:"content,omitempty"`
	AuthorId int64  `gorm:"index" bson:"author_id,omitempty"`
	//AuthorId int64  `gorm:"index:aid_ctime"`
	//Ctime    int64  `gorm:"index:aid_ctime"`
	Ctime  int64 `bson:"ctime,omitempty"`
	Utime  int64 `bson:"utime,omitempty"`
	Status uint8 `bson:"status,omitempty"`
}

// PublishedArticle 线上库
type PublishedArticle struct {
	Article
}
