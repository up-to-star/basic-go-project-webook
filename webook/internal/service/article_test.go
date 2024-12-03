package service

import (
	"basic-project/webook/internal/domain"
	"basic-project/webook/internal/repository/article"
	repomocks "basic-project/webook/internal/repository/mocks"
	"context"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func Test_articleService_Publish(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) article.ArticleRepository
		art     domain.Article
		wantId  int64
		wantErr error
	}{
		{
			name: "发表成功",
			mock: func(ctrl *gomock.Controller) article.ArticleRepository {
				repo := repomocks.NewMockArticleRepository(ctrl)
				repo.EXPECT().Sync(gomock.Any(), gomock.Any()).Return(int64(123), nil)
				return repo
			},
			art: domain.Article{
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantId:  123,
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repo := tc.mock(ctrl)
			svc := NewArticleService(repo)
			artId, err := svc.Publish(context.Background(), tc.art)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantId, artId)
		})
	}
}
