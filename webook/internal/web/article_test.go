package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/basic-go-project-webook/webook/internal/domain"
	"github.com/basic-go-project-webook/webook/internal/service"
	svcmocks "github.com/basic-go-project-webook/webook/internal/service/mocks"
	ijwt "github.com/basic-go-project-webook/webook/internal/web/jwt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestArticleHandle_Publish(t *testing.T) {
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) service.ArticleService
		reqBody  string
		wantCode int
		wantRes  Result
	}{
		{
			name: "新建并发表",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "new article and publish",
					Content: "this is the content",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)
				return svc
			},
			reqBody: `
{
	"title": "new article and publish",
	"content": "this is the content"
}`,
			wantCode: http.StatusOK,
			wantRes: Result{
				Code: 0,
				Msg:  "OK",
				Data: float64(1),
			},
		},
		{
			name: "publish失败",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "new article and publish",
					Content: "this is the content",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(0), errors.New("publish error"))
				return svc
			},
			reqBody: `
{
	"title": "new article and publish",
	"content": "this is the content"
}`,
			wantCode: http.StatusOK,
			wantRes: Result{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			articleService := tc.mock(ctrl)
			cmd := InitRedis()
			handle := NewArticleHandle(articleService, ijwt.NewRedisJwtHandler(cmd))
			server := gin.Default()
			handle.RegisterRoutes(server)
			tokenStr := generateToken(123)
			req, err := http.NewRequest(http.MethodPost, "/articles/publish", bytes.NewBuffer([]byte(tc.reqBody)))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+tokenStr)
			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, req)
			assert.Equal(t, tc.wantCode, recorder.Code)
			var webRes Result
			err = json.NewDecoder(recorder.Body).Decode(&webRes)
			require.NoError(t, err)
			assert.Equal(t, tc.wantRes, webRes)
		})
	}
}

func generateToken(uid int64) string {
	claims := ijwt.UserClaims{
		Uid: uid,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString(ijwt.AtKey)
	return tokenStr
}

func InitRedis() redis.Cmdable {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6380",
	})
}
