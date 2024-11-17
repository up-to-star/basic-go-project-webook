package repository

import (
	"basic-project/webook/internal/domain"
	"basic-project/webook/internal/repository/cache"
	"basic-project/webook/internal/repository/dao"
	"context"
	"database/sql"
	"time"
)

var (
	ErrUserDuplicateEmail = dao.ErrUserDuplicate
	ErrUserNotFound       = dao.ErrUserNotFount
)

type UserRepository struct {
	dao   *dao.UserDAO
	cache *cache.UserCache
}

func NewUserRepository(dao *dao.UserDAO, cache *cache.UserCache) *UserRepository {
	return &UserRepository{
		dao:   dao,
		cache: cache,
	}
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) error {
	return r.dao.Insert(ctx, r.domainToEntity(user))
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	u := r.entityToDomain(user)
	return u, nil
}

func (r *UserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	user, err := r.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(user), nil
}

func (r *UserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
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
	err = r.cache.Set(ctx, user)
	if err != nil {
		// 日志监控
	}
	return user, err
}

func (r *UserRepository) entityToDomain(ud dao.User) domain.User {
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

func (r *UserRepository) domainToEntity(u domain.User) dao.User {
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
