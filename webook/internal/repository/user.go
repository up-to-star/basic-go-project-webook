package repository

import (
	"basic-project/webook/internal/domain"
	"basic-project/webook/internal/repository/cache"
	"basic-project/webook/internal/repository/dao"
	"context"
	"database/sql"
	"github.com/gin-gonic/gin"
	"time"
)

var (
	ErrUserDuplicateEmail = dao.ErrUserDuplicate
	ErrUserNotFound       = dao.ErrUserNotFount
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	FindById(ctx context.Context, id int64) (domain.User, error)
	UpdateById(ctx *gin.Context, user domain.User) error
}

type CachedUserRepository struct {
	dao   dao.UserDAO
	cache cache.UserCache
}

func NewUserRepository(dao dao.UserDAO, cache cache.UserCache) UserRepository {
	return &CachedUserRepository{
		dao:   dao,
		cache: cache,
	}
}

func (r *CachedUserRepository) Create(ctx context.Context, user domain.User) error {
	return r.dao.Insert(ctx, r.domainToEntity(user))
}

func (r *CachedUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	u := r.entityToDomain(user)
	return u, nil
}

func (r *CachedUserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	user, err := r.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(user), nil
}

func (r *CachedUserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	user, err := r.cache.Get(ctx, id)
	if err == nil {
		// 缓存里有数据
		return user, nil
	}
	// 缓存里没有数据
	//if errors.Is(err, cache.ErrKeyNotExists) {
	//
	//}
	u, err := r.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	user = r.entityToDomain(u)
	_ = r.cache.Set(ctx, user)
	//if err != nil {
	//	// 日志监控
	//}
	return user, nil
}

func (r *CachedUserRepository) entityToDomain(ud dao.User) domain.User {
	return domain.User{
		Id:       ud.Id,
		Email:    ud.Email.String,
		Password: ud.Password,
		Nickname: ud.Nickname,
		Phone:    ud.Phone.String,
		Birthday: time.UnixMilli(ud.Birthday),
		Ctime:    time.UnixMilli(ud.Ctime),
		Utime:    time.UnixMilli(ud.Utime),
		AboutMe:  ud.AboutMe,
	}
}

func (r *CachedUserRepository) domainToEntity(u domain.User) dao.User {
	return dao.User{
		Id: u.Id,
		Email: sql.NullString{
			String: u.Email,
			Valid:  u.Email != "",
		},
		Password: u.Password,
		Nickname: u.Nickname,
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  u.Phone != "",
		},
		AboutMe:  u.AboutMe,
		Birthday: u.Birthday.UnixMilli(),
		Ctime:    u.Ctime.UnixMilli(),
		Utime:    u.Utime.UnixMilli(),
	}
}

func (r *CachedUserRepository) UpdateById(ctx *gin.Context, user domain.User) error {
	_, err := r.cache.Get(ctx, user.Id)
	if err == nil {
		_ = r.cache.Del(ctx, user.Id)
	}
	return r.dao.UpdateById(ctx, r.domainToEntity(user))
}
