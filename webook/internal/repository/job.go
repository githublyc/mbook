package repository

import (
	"context"
	"mbook/webook/internal/domain"
	"mbook/webook/internal/repository/dao"
	"time"
)

type CronJobRepository interface {
	Preempt(ctx context.Context) (domain.Job, error)
	Release(ctx context.Context, jid int64) error
	UpdateUtime(ctx context.Context, jid int64) error
	UpdateNextTime(ctx context.Context, jid int64, time time.Time) error
}
type PreemptJobRepository struct {
	dao dao.JobDAO
}

func (p *PreemptJobRepository) UpdateNextTime(ctx context.Context, jid int64, time time.Time) error {
	return p.dao.UpdateNextTime(ctx, jid, time)
}

func (p *PreemptJobRepository) UpdateUtime(ctx context.Context, jid int64) error {
	return p.dao.UpdateUtime(ctx, jid)
}

func (p *PreemptJobRepository) Preempt(ctx context.Context) (domain.Job, error) {
	j, err := p.dao.Preempt(ctx)
	return domain.Job{
		Id:         j.Id,
		Expression: j.Expression,
		Executor:   j.Executor,
		Name:       j.Name,
	}, err
}

func (p *PreemptJobRepository) Release(ctx context.Context, jid int64) error {
	return p.dao.Release(ctx, jid)
}
