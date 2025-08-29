package job

import (
	"context"
	rlock "github.com/gotomicro/redis-lock"
	"mbook/webook/internal/service"
	"mbook/webook/pkg/logger"
	"sync"
	"time"
)

type RankingJob struct {
	svc     service.RankingService
	l       logger.LoggerV1
	timeout time.Duration
	client  *rlock.Client
	key     string

	localLock *sync.Mutex
	lock      *rlock.Lock
}

func NewRankingJob(svc service.RankingService, l logger.LoggerV1,
	client *rlock.Client, timeout time.Duration) *RankingJob {
	return &RankingJob{
		svc:       svc,
		l:         l,
		key:       "job:ranking",
		timeout:   timeout,
		client:    client,
		localLock: &sync.Mutex{},
	}
}

func (r *RankingJob) Name() string {
	return "RankingJob"
}

func (r *RankingJob) Run() error {
	r.localLock.Lock()
	lock := r.lock
	if lock == nil {
		// 抢分布式锁
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
		defer cancel()
		lock, err := r.client.Lock(ctx, r.key, r.timeout,
			&rlock.FixIntervalRetry{
				//失败后休息多久再试
				Interval: time.Millisecond * 100,
				Max:      3,
				// 重试的超时
			}, time.Second)
		if err != nil {
			r.localLock.Unlock()
			r.l.Warn("获取分布式锁失败", logger.Error(err))
			return nil
		}
		r.lock = lock
		r.localLock.Unlock()
		//续约
		go func() {
			// 并不是非得一半就续约
			er := lock.AutoRefresh(r.timeout/2, r.timeout)
			if er != nil {
				// 续约失败了
				// 你也没办法中断当下正在调度的热榜计算（如果有）
				r.localLock.Lock()
				r.lock = nil
				r.localLock.Unlock()
			}
		}()
	}
	// 这边就是你拿到了锁,执行业务逻辑（计算热榜）
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	return r.svc.TopN(ctx)
}

func (r *RankingJob) Close() error {
	r.localLock.Lock()
	lock := r.lock
	r.localLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return lock.Unlock(ctx)
}

//func (r *RankingJob) Run() error {
//	ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
//	defer cancel()
//	lock, err := r.client.Lock(ctx, r.key, r.timeout, &rlock.FixIntervalRetry{
//		Interval: time.Millisecond * 100,
//		Max:      3,
//		// 重试的超时
//	}, time.Second)
//	if err != nil {
//		return err
//	}
//	defer func() {
//		// 解锁
//		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
//		defer cancel()
//		er := lock.Unlock(ctx)
//		if er != nil {
//			r.l.Error("ranking job释放分布式锁失败", logger.Error(er))
//		}
//	}()
//	ctx, cancel = context.WithTimeout(context.Background(), r.timeout)
//	defer cancel()
//	return r.svc.TopN(ctx)
//}
