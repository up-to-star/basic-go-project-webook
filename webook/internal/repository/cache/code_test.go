package cache

import (
	"context"
	"errors"
	"github.com/basic-go-project-webook/webook/internal/repository/cache/redismocks"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestRedisCodeCache_Set(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) redis.Cmdable
		biz     string
		phone   string
		code    string
		wantErr error
	}{
		{
			name: "验证码存储成功",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(0))
				cmd.EXPECT().Eval(gomock.Any(), luaSetCode, []string{"phone_code:login:15716604112"}, []any{"123456"}).Return(res)
				return cmd
			},
			biz:     "login",
			phone:   "15716604112",
			code:    "123456",
			wantErr: nil,
		},
		{
			name: "redis错误",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetErr(errors.New("mock redis error"))
				// res.SetVal(int64(0))
				cmd.EXPECT().Eval(gomock.Any(), luaSetCode, []string{"phone_code:login:15716604112"}, []any{"123456"}).Return(res)
				return cmd
			},
			biz:     "login",
			phone:   "15716604112",
			code:    "123456",
			wantErr: errors.New("mock redis error"),
		},
		{
			name: "发送太频繁",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(-1))
				cmd.EXPECT().Eval(gomock.Any(), luaSetCode, []string{"phone_code:login:15716604112"}, []any{"123456"}).Return(res)
				return cmd
			},
			biz:     "login",
			phone:   "15716604112",
			code:    "123456",
			wantErr: ErrCodeSendTooMany,
		},
		{
			name: "验证码存储成功",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				// res.SetErr(errors.New("system error"))
				res.SetVal(int64(-10))
				cmd.EXPECT().Eval(gomock.Any(), luaSetCode, []string{"phone_code:login:15716604112"}, []any{"123456"}).Return(res)
				return cmd
			},
			biz:     "login",
			phone:   "15716604112",
			code:    "123456",
			wantErr: errors.New("system error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := NewCodeCache(tc.mock(ctrl))
			err := c.Set(context.Background(), tc.biz, tc.phone, tc.code)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
