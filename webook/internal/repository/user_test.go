package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/basic-go-project-webook/webook/internal/domain"
	"github.com/basic-go-project-webook/webook/internal/repository/cache"
	cachemocks "github.com/basic-go-project-webook/webook/internal/repository/cache/mocks"
	"github.com/basic-go-project-webook/webook/internal/repository/dao"
	daomocks "github.com/basic-go-project-webook/webook/internal/repository/dao/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestCachedUserRepository_FindById(t *testing.T) {
	now := time.Now()
	now = time.UnixMilli(now.UnixMilli())
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache)
		id       int64
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "缓存未命中，查询成功",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), int64(123)).Return(domain.User{}, cache.ErrKeyNotExists)
				d := daomocks.NewMockUserDAO(ctrl)
				d.EXPECT().FindById(gomock.Any(), int64(123)).Return(dao.User{
					Id: 123,
					Email: sql.NullString{
						String: "123@qq.com",
						Valid:  true,
					},
					Password: "hello#world123",
					Phone: sql.NullString{
						String: "12345678901",
						Valid:  true,
					},
					Birthday: now.UnixMilli(),
					Ctime:    now.UnixMilli(),
					Utime:    now.UnixMilli(),
				}, nil)
				c.EXPECT().Set(gomock.Any(), gomock.Any()).Return(nil)
				return d, c
			},
			id: 123,
			wantUser: domain.User{
				Id:       123,
				Email:    "123@qq.com",
				Password: "hello#world123",
				Phone:    "12345678901",
				Birthday: now,
				Ctime:    now,
				Utime:    now,
			},
			wantErr: nil,
		},
		{
			name: "缓存命中",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), int64(123)).Return(domain.User{
					Id:       123,
					Email:    "123@qq.com",
					Password: "hello#world123",
					Phone:    "12345678901",
					Birthday: now,
					Ctime:    now,
					Utime:    now,
				}, nil)
				d := daomocks.NewMockUserDAO(ctrl)

				return d, c
			},
			id: 123,
			wantUser: domain.User{
				Id:       123,
				Email:    "123@qq.com",
				Password: "hello#world123",
				Phone:    "12345678901",
				Birthday: now,
				Ctime:    now,
				Utime:    now,
			},
			wantErr: nil,
		},
		{
			name: "缓存未命中，查询失败",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), int64(123)).Return(domain.User{}, cache.ErrKeyNotExists)
				d := daomocks.NewMockUserDAO(ctrl)
				d.EXPECT().FindById(gomock.Any(), int64(123)).Return(dao.User{}, errors.New("mock db error"))
				return d, c
			},
			id:       123,
			wantUser: domain.User{},
			wantErr:  errors.New("mock db error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ud, uc := tc.mock(ctrl)
			repo := NewUserRepository(ud, uc)
			user, err := repo.FindById(context.Background(), tc.id)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}
}
