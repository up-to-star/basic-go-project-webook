package repository

import (
	"context"
	"github.com/basic-go-project-webook/webook/internal/domain"
	"github.com/basic-go-project-webook/webook/internal/repository/dao"
	"time"
)

type CronJobRepository interface {
	Preempt(ctx context.Context) (domain.Job, error)
	UpdateTime(ctx context.Context, id int64) error
	Release(ctx context.Context, id int64) error
	UpdateNextTime(ctx context.Context, id int64, next time.Time) error
}

type PreemptJobRepository struct {
	dao dao.JobDAO
}

func (p *PreemptJobRepository) UpdateNextTime(ctx context.Context, id int64, next time.Time) error {
	return p.dao.UpdateNextTime(ctx, id, next)
}

func (p *PreemptJobRepository) Release(ctx context.Context, id int64) error {
	return p.dao.Release(ctx, id)
}

func (p *PreemptJobRepository) UpdateTime(ctx context.Context, id int64) error {
	return p.dao.UpdateTime(ctx, id)
}

func (p *PreemptJobRepository) Preempt(ctx context.Context) (domain.Job, error) {
	j, err := p.dao.Preempt(ctx)
	if err != nil {
		return domain.Job{}, err
	}
	return domain.Job{
		Id:         j.Id,
		Name:       j.Name,
		Expression: j.Expression,
		Cfg:        j.Cfg,
	}, nil
}

func NewPreemptJobRepository(dao dao.JobDAO) CronJobRepository {
	return &PreemptJobRepository{
		dao: dao,
	}
}
