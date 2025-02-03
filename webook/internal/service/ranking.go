package service

import (
	"basic-project/webook/internal/domain"
	"basic-project/webook/internal/repository"
	"context"
	"errors"
	"github.com/ecodeclub/ekit/queue"
	"math"
	"time"
)

type RankingService interface {
	TopN(ctx context.Context) error
}

type BatchRankingService struct {
	artSvc    ArticleService
	intrSvc   InteractiveService
	repo      repository.RankingRepository
	batchSize int
	n         int
	scoreFunc func(t time.Time, likeCnt int64) float64
}

func NewBatchRankingService(artSvc ArticleService, intrSvc InteractiveService, repo repository.RankingRepository) RankingService {
	return &BatchRankingService{
		artSvc:    artSvc,
		intrSvc:   intrSvc,
		repo:      repo,
		batchSize: 100,
		n:         100,
		scoreFunc: func(t time.Time, likeCnt int64) float64 {
			sec := time.Since(t).Seconds()
			return float64(likeCnt-1) / math.Pow(sec+2, 1.5)
		},
	}
}

func (svc *BatchRankingService) TopN(ctx context.Context) error {
	arts, err := svc.topN(ctx)
	if err != nil {
		return err
	}
	// 存入缓存中
	return svc.repo.ReplaceTopN(ctx, arts)
}

func (svc *BatchRankingService) topN(ctx context.Context) ([]domain.Article, error) {
	start := time.Now()
	offset := 0
	ddl := start.Add(-7 * 24 * time.Hour)
	type Score struct {
		art   domain.Article
		score float64
	}
	topN := queue.NewPriorityQueue[Score](svc.n, func(src, dst Score) int {
		if src.score > dst.score {
			return 1
		} else if src.score == dst.score {
			return 0
		} else {
			return -1
		}
	})

	for {
		arts, err := svc.artSvc.ListPub(ctx, start, offset, svc.batchSize)
		if err != nil {
			return nil, err
		}

		ids := make([]int64, len(arts))
		for i, art := range arts {
			ids[i] = art.Id
		}

		intrMap, err := svc.intrSvc.GetByIds(ctx, "article", ids)
		if err != nil {
			return nil, err
		}

		for _, art := range arts {
			intr, ok := intrMap[art.Id]
			if !ok {
				continue
			}
			score := svc.scoreFunc(art.Utime, intr.LikeCnt)
			ele := Score{
				art:   art,
				score: score,
			}
			err = topN.Enqueue(ele)
			if errors.Is(err, queue.ErrOutOfCapacity) {
				minEle, _ := topN.Dequeue()
				if minEle.score < score {
					_ = topN.Enqueue(ele)
				} else {
					_ = topN.Enqueue(minEle)
				}
			}
		}
		offset += svc.batchSize
		if len(arts) < svc.batchSize || arts[len(arts)-1].Utime.Before(ddl) {
			break
		}
	}
	res := make([]domain.Article, topN.Len())
	for i := topN.Len() - 1; i >= 0; i-- {
		ele, _ := topN.Dequeue()
		res[i] = ele.art
	}
	return res, nil
}
