package service

import (
	"basic-project/webook/internal/domain"
	"basic-project/webook/internal/repository"
	"context"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (svc *UserService) Signup(ctx context.Context, user domain.User) error {
	return svc.repo.Create(ctx, user)
}
