package service

import (
	"context"
	"mbook/webook/internal/domain"
	"mbook/webook/internal/repository"
	"mbook/webook/pkg/logger"
	"time"
)

type CronJobService interface {
	Preempt(ctx context.Context) (domain.Job, error)
	ResetNextTime(ctx context.Context, j domain.Job) error
	// 暴露 job 的增删改查方法
}

type cronJobService struct {
	repo            repository.CronJobRepository
	l               logger.LoggerV1
	refreshInterval time.Duration
}

func (c *cronJobService) ResetNextTime(ctx context.Context, j domain.Job) error {
	nextTime := j.NextTime()
	return c.repo.UpdateNextTime(ctx, j.Id, nextTime)
}

func (c *cronJobService) Preempt(ctx context.Context) (domain.Job, error) {
	j, err := c.repo.Preempt(ctx)
	if err != nil {
		return domain.Job{}, err
	}
	ticker := time.NewTicker(c.refreshInterval)
	go func() {
		for range ticker.C {
			c.refresh(j.Id)
		}
	}()
	j.CancelFunc = func() {
		ticker.Stop()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err := c.repo.Release(ctx, j.Id)
		if err != nil {
			c.l.Error("释放 job 失败",
				logger.Error(err),
				logger.Int64("job_id", j.Id))
		}
	}
	return j, nil
}

func (c *cronJobService) refresh(jid int64) {
	// 本质上就是更新一下更新时间
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := c.repo.UpdateUtime(ctx, jid)
	if err != nil {
		c.l.Error("续约失败", logger.Error(err), logger.Int64("job_id", jid))
	}
}
