package grpc

import (
	"context"
	"github.com/ecodeclub/ekit/queue"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
	"sync/atomic"
	"time"
)

type CounterLimiter struct {
	cnt       atomic.Int32
	threshold int32
}

func (c *CounterLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp any, err error) {
		// 请求进来，先占坑
		cnt := c.cnt.Add(1)
		defer func() {
			c.cnt.Add(-1)
		}()
		if cnt <= c.threshold {
			resp, err = handler(ctx, req)
			// 返回了响应
			return
		}
		return nil, status.Errorf(codes.ResourceExhausted,
			"限流")
	}
}

type FixedWindowLimiter struct {
	window          time.Duration
	lastWindowStart time.Time
	cnt             int
	threshold       int
	lock            sync.Mutex
}

func NewFixedWindowLimiter(window time.Duration,
	threshold int) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		window:          window,
		lastWindowStart: time.Now(),
		cnt:             0,
		threshold:       threshold,
	}
}

func (f *FixedWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp any, err error) {
		f.lock.Lock()

		now := time.Now()
		// 换新窗口
		if now.After(f.lastWindowStart.Add(f.window)) {
			f.cnt = 0
			f.lastWindowStart = now
		}
		cnt := f.cnt + 1
		f.lock.Unlock()
		if cnt <= f.threshold {
			resp, err = handler(ctx, req)
			return
		}
		return nil, status.Errorf(codes.ResourceExhausted,
			"限流")
	}
}

type SlidingWindowLimiter struct {
	window time.Duration
	// 请求到来的时间戳
	// 时间戳最小的在队头
	queue     queue.PriorityQueue[time.Time]
	lock      sync.Mutex
	threshold int
}

func (s *SlidingWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp any, err error) {
		s.lock.Lock()
		now := time.Now()

		// 快路径检测
		if s.queue.Len() < s.threshold {
			_ = s.queue.Enqueue(now)
			s.lock.Unlock()
			resp, err = handler(ctx, req)
			return
		}

		// 我先考虑队列里面的时间戳是不是都在我的窗口范围内
		windowStart := now.Add(-s.window)
		for {
			first, _ := s.queue.Peek()
			if first.Before(windowStart) {
				// 把第一个元素删了
				_, _ = s.queue.Dequeue()
			} else {
				break
			}
		}
		if s.queue.Len() < s.threshold {
			_ = s.queue.Enqueue(now)
			s.lock.Unlock()
			resp, err = handler(ctx, req)
			return
		}
		s.lock.Unlock()
		return nil, status.Errorf(codes.ResourceExhausted,
			"限流")
	}
}

type TokenBucketLimiter struct {
	// 隔多久产生一个令牌
	interval  time.Duration
	buckets   chan struct{}
	closeCh   chan struct{}
	closeOnce sync.Once
}

func (t *TokenBucketLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	ticker := time.NewTicker(t.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				select {
				case t.buckets <- struct{}{}:
				default:
					// bucket 满了
				}
			case <-t.closeCh:
				return
			}
		}
	}()
	return func(ctx context.Context, req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp any, err error) {
		select {
		case <-t.buckets:
			return handler(ctx, req)
		// 做法1
		default:
			return nil, status.Errorf(codes.ResourceExhausted,
				"限流")
			// 做法2
			//case <-ctx.Done():
			//	return nil, ctx.Err()
		}

	}
}
func (t *TokenBucketLimiter) Close() error {
	t.closeOnce.Do(func() {
		close(t.closeCh)
	})
	return nil
}

type LeakyBucketLimiter struct {
	// 隔多久产生一个令牌
	interval  time.Duration
	closeCh   chan struct{}
	closeOnce sync.Once
}

func (l *LeakyBucketLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	ticker := time.NewTicker(l.interval)
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		select {
		case <-ticker.C:
			return handler(ctx, req)
		case <-l.closeCh:
			// 限流器已经关了
			return nil,
				status.Errorf(codes.ResourceExhausted, "限流")
		//做法1
		default:
			return nil,
				status.Errorf(codes.ResourceExhausted, "限流")
			// 做法2
			//case <-ctx.Done():
			//	return nil, ctx.Err()
		}
	}
}

func (l *LeakyBucketLimiter) Close() error {
	l.closeOnce.Do(func() {
		close(l.closeCh)
	})
	return nil
}
