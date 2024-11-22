package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

var (
	ErrUserDuplicate = errors.New("账号已经被注册过了")
	ErrUserNotFount  = gorm.ErrRecordNotFound
)

type UserDAO interface {
	Insert(ctx context.Context, user User) error
	FindByEmail(ctx context.Context, email string) (User, error)
	FindById(ctx context.Context, id int64) (User, error)
	FindByPhone(ctx context.Context, phone string) (User, error)
	UpdateById(ctx *gin.Context, user User) error
	FindByWechat(ctx *gin.Context, openId string) (User, error)
}

type GORMUserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) UserDAO {
	return &GORMUserDAO{
		db: db,
	}
}

func (dao *GORMUserDAO) Insert(ctx context.Context, user User) error {
	now := time.Now().UnixMilli()
	user.Utime = now
	user.Ctime = now
	err := dao.db.WithContext(ctx).Create(&user).Error
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		const uniqueConflictsErrNo uint16 = 1062
		if mysqlErr.Number == uniqueConflictsErrNo {
			return ErrUserDuplicate
		}
	}
	return err
}

func (dao *GORMUserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := dao.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	return user, err
}

func (dao *GORMUserDAO) FindById(ctx context.Context, id int64) (User, error) {
	var user User
	err := dao.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	return user, err
}

func (dao *GORMUserDAO) FindByPhone(ctx context.Context, phone string) (User, error) {
	var user User
	err := dao.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error
	return user, err
}

func (dao *GORMUserDAO) UpdateById(ctx *gin.Context, user User) error {
	now := time.Now().UnixMilli()
	user.Utime = now
	return dao.db.WithContext(ctx).Model(&User{}).Where("id = ?", user.Id).Updates(user).Error
}

func (dao *GORMUserDAO) FindByWechat(ctx *gin.Context, openId string) (User, error) {
	var user User
	err := dao.db.WithContext(ctx).Where("wechat_open_id = ?", openId).First(&user).Error
	return user, err
}

// User 直接对应数据库表
type User struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 唯一索引允许有多个null, 不允许有多个 ""
	Email         sql.NullString `gorm:"type:varchar(255);unique"`
	Password      string         `gorm:"type:varchar(255)"`
	Phone         sql.NullString `gorm:"type:char(11);unique"`
	WechatUnionID sql.NullString `gorm:"type:varchar(255)"`
	WechatOpenID  sql.NullString `gorm:"type:varchar(255);unique"`
	Nickname      string         `gorm:"type:varchar(128)"`
	AboutMe       string         `gorm:"type:varchar(4096)"`
	Birthday      int64
	Ctime         int64
	Utime         int64
}
