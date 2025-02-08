package service

import (
	"context"
	"errors"
	"github.com/basic-go-project-webook/webook/internal/domain"
	"github.com/basic-go-project-webook/webook/internal/repository"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserDuplicateEmail    = repository.ErrUserDuplicateEmail
	ErrInvalidUserOrPassword = errors.New("邮箱或密码错误")
)

type UserService interface {
	Signup(ctx context.Context, user domain.User) error
	Login(ctx context.Context, user domain.User) (domain.User, error)
	Profile(ctx context.Context, id int64) (domain.User, error)
	FindOrCreate(ctx *gin.Context, phone string) (domain.User, error)
	Edit(ctx *gin.Context, user domain.User) error
	FindOrCreateByWechat(ctx *gin.Context, info domain.WechatInfo) (domain.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (svc *userService) Signup(ctx context.Context, user domain.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hash)
	return svc.repo.Create(ctx, user)
}

func (svc *userService) Login(ctx context.Context, user domain.User) (domain.User, error) {
	u, err := svc.repo.FindByEmail(ctx, user.Email)
	if errors.Is(err, repository.ErrUserNotFound) {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	// 比较密码
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(user.Password))
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, nil
}

func (svc *userService) Profile(ctx context.Context, id int64) (domain.User, error) {
	user, err := svc.repo.FindById(ctx, id)
	return user, err
}

func (svc *userService) FindOrCreate(ctx *gin.Context, phone string) (domain.User, error) {
	user, err := svc.repo.FindByPhone(ctx, phone)
	if !errors.Is(err, repository.ErrUserNotFound) {
		return user, err
	}
	u := domain.User{
		Phone: phone,
	}
	err = svc.repo.Create(ctx, u)
	if err != nil {
		return u, err
	}
	return svc.repo.FindByPhone(ctx, phone)
}

func (svc *userService) FindOrCreateByWechat(ctx *gin.Context, info domain.WechatInfo) (domain.User, error) {
	user, err := svc.repo.FindByWechat(ctx, info.OpenId)
	if !errors.Is(err, repository.ErrUserNotFound) {
		return user, err
	}
	u := domain.User{
		WechatInfo: info,
	}
	err = svc.repo.Create(ctx, u)
	if err != nil {
		return u, err
	}
	return svc.repo.FindByWechat(ctx, info.OpenId)
}

func (svc *userService) Edit(ctx *gin.Context, user domain.User) error {
	return svc.repo.UpdateById(ctx, user)
}
