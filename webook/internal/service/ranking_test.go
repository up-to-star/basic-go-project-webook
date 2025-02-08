package service

import (
	domain2 "github.com/basic-go-project-webook/webook/interactive/domain"
	"github.com/basic-go-project-webook/webook/interactive/service"
	"github.com/basic-go-project-webook/webook/internal/domain"
	svcmocks "github.com/basic-go-project-webook/webook/internal/service/mocks"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestRankingTopN(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) (ArticleService, service.InteractiveService)
		wantErr  error
		wantArts []domain.Article
	}{
		{
			name: "计算成功",
			mock: func(ctrl *gomock.Controller) (ArticleService, service.InteractiveService) {
				artSvc := svcmocks.NewMockArticleService(ctrl)
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 0, 3).Return([]domain.Article{
					{Id: 1, Ctime: now, Utime: now},
					{Id: 2, Ctime: now, Utime: now},
					{Id: 3, Ctime: now, Utime: now},
				}, nil)
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 3, 3).Return([]domain.Article{}, nil)
				intrSvc := svcmocks.NewMockInteractiveService(ctrl)
				intrSvc.EXPECT().GetByIds(gomock.Any(), "article", []int64{1, 2, 3}).Return(map[int64]domain2.Interactive{
					1: {BizId: 1, LikeCnt: 1},
					2: {BizId: 2, LikeCnt: 2},
					3: {BizId: 3, LikeCnt: 3},
				}, nil)
				intrSvc.EXPECT().GetByIds(gomock.Any(), "article", []int64{}).Return(map[int64]domain2.Interactive{}, nil)

				return artSvc, intrSvc
			},
			wantArts: []domain.Article{
				{Id: 3, Ctime: now, Utime: now},
				{Id: 2, Ctime: now, Utime: now},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			//artSvc, intrSvc := tc.mock(ctrl)
			//svc := NewBatchRankingService(artSvc, intrSvc)
			//svc.n = 2
			//svc.batchSize = 3
			//svc.scoreFunc = func(t time.Time, likeCnt int64) float64 {
			//	return float64(likeCnt)
			//}
			//arts, err := svc.topN(context.Background())
			//t.Log("arts: ", arts)
			//assert.Equal(t, tc.wantErr, err)
			//assert.Equal(t, tc.wantArts, arts)
		})
	}
}
