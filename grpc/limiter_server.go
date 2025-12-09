package grpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"mbook/webook/pkg/limiter"
)

// LimiterUserServiceServer 只是转发不是实现，所以不需要组合UnimplementedUserServiceServer
type LimiterUserServiceServer struct {
	limiter limiter.Limiter
	UserServiceServer
}

// 装饰器
// 比BuildServerUnaryInterceptorBiz更好，不用在其他业务前判断是否限流
func (s *LimiterUserServiceServer) GetByID(ctx context.Context,
	req *GetByIDRequest) (*GetByIDResponse, error) {
	key := fmt.Sprintf("limiter:user:get_by_id:%d", req.Id)
	limited, err := s.limiter.Limit(ctx, key)
	if err != nil {
		// 有保守的做法，也有激进的做法
		// 这个是保守的做法
		return nil,
			status.Errorf(codes.ResourceExhausted, "限流")

	}
	if limited {
		return nil,
			status.Errorf(codes.ResourceExhausted, "限流")
	}
	return s.UserServiceServer.GetByID(ctx, req)
}
