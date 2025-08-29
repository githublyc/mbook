package job

import (
	"context"
	"fmt"
	"golang.org/x/sync/semaphore"
	"mbook/webook/internal/domain"
	"mbook/webook/internal/service"
	"mbook/webook/pkg/logger"
	"time"
)

// Executor 执行器，任务执行器
type Executor interface {
	Name() string
	// Exec ctx 这个是全局控制，Executor 的实现者注意要正确处理 ctx 超时或者取消
	Exec(ctx context.Context, j domain.Job) error
}

// LocalFuncExecutor 调用本地方法的
type LocalFuncExecutor struct {
	funcs map[string]func(ctx context.Context, j domain.Job) error
}

func NewLocalFuncExecutor() *LocalFuncExecutor {
	return &LocalFuncExecutor{
		funcs: make(map[string]func(ctx context.Context, j domain.Job) error),
	}
}

func (l *LocalFuncExecutor) Name() string {
	return "Local"
}

func (l *LocalFuncExecutor) Exec(ctx context.Context, j domain.Job) error {
	fn, ok := l.funcs[j.Name]
	if !ok {
		return fmt.Errorf("未注册本地方法 %s", j.Name)
	}
	return fn(ctx, j)
}
func (l *LocalFuncExecutor) RegisterFunc(name string,
	fn func(ctx context.Context, j domain.Job) error) {
	l.funcs[name] = fn
}

type Scheduler struct {
	dbTimeout time.Duration
	svc       service.CronJobService
	executors map[string]Executor
	l         logger.LoggerV1

	limiter *semaphore.Weighted
}

func NewScheduler(svc service.CronJobService, l logger.LoggerV1) *Scheduler {
	return &Scheduler{
		svc:       svc,
		l:         l,
		dbTimeout: time.Second,
		limiter:   semaphore.NewWeighted(100),
		executors: make(map[string]Executor),
	}
}

func (s *Scheduler) RegisterExecutor(exec Executor) {
	s.executors[exec.Name()] = exec
}

func (s *Scheduler) Schedule(ctx context.Context) {
	for {
		//放弃调度了
		if ctx.Err() != nil {
			return
		}
		err := s.limiter.Acquire(ctx, 1)
		if err != nil {
			return
		}
		dbCtx, cancel := context.WithTimeout(ctx, s.dbTimeout)
		j, err := s.svc.Preempt(dbCtx)
		cancel()
		if err != nil {
			continue
		}

		// 肯定要调度执行 j
		exec, ok := s.executors[j.Executor]
		if !ok {
			// 可以直接中断了，也可以下一轮
			s.l.Error("找不到执行器",
				logger.Int64("jid", j.Id),
				logger.String("executor", j.Executor))
			continue
		}

		go func() {
			defer func() {
				s.limiter.Release(1)
				// 这边要释放掉
				j.CancelFunc()
			}()
			err1 := exec.Exec(ctx, j)
			if err1 != nil {
				s.l.Error("执行任务失败",
					logger.Int64("jid", j.Id),
					logger.Error(err1))
				return
			}
			err1 = s.svc.ResetNextTime(ctx, j)
			if err1 != nil {
				s.l.Error("重置下次执行时间失败",
					logger.Int64("jid", j.Id),
					logger.Error(err1))
			}
		}()
	}
}
