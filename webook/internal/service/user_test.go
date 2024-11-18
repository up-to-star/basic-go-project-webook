package service

import (
	"basic-project/webook/internal/domain"
	"basic-project/webook/internal/repository"
	repomocks "basic-project/webook/internal/repository/mocks"
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"testing"
	"time"
)

func Test_userService_Login(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) repository.UserRepository
		email    string
		password string
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "登录成功",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), gomock.Any()).Return(domain.User{
					Email:    "test@test.com",
					Password: "$2a$10$Qc10YngGuMSpvpbnuto09.YZMeuwgzoIXdKtY62vx3aFzIWSLkj7O",
					Phone:    "15716604112",
					Ctime:    now,
				}, nil)
				return repo
			},
			email:    "test@test.com",
			password: "hello#world123",
			wantUser: domain.User{
				Email:    "test@test.com",
				Password: "$2a$10$Qc10YngGuMSpvpbnuto09.YZMeuwgzoIXdKtY62vx3aFzIWSLkj7O",
				Phone:    "15716604112",
				Ctime:    now,
			},
			wantErr: nil,
		},
		{
			name: "用户不存在",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), gomock.Any()).Return(domain.User{}, repository.ErrUserNotFound)
				return repo
			},
			email:    "test@test.com",
			password: "hello#world123",
			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), gomock.Any()).Return(domain.User{}, errors.New("DB error"))
				return repo
			},
			email:    "test@test.com",
			password: "hello#world123",
			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
		{
			name: "密码错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), gomock.Any()).Return(domain.User{}, ErrInvalidUserOrPassword)
				return repo
			},
			email:    "test@test.com",
			password: "hello#wo23432rld123",
			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 具体的测试代码
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			svc := NewUserService(tc.mock(ctrl))
			user, err := svc.Login(context.Background(), domain.User{Email: tc.email, Password: tc.password})
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}
}

func TestEncrypted(t *testing.T) {
	res, err := bcrypt.GenerateFromPassword([]byte("hello#world123"), bcrypt.DefaultCost)
	if err == nil {
		t.Log(string(res))
	}
}
