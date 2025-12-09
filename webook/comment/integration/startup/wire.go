//go:build wireinject

package startup

import (
	"github.com/google/wire"
	grpc2 "mbook/webook/comment/grpc"
	"mbook/webook/comment/repository"
	"mbook/webook/comment/repository/dao"
	"mbook/webook/comment/service"
	"mbook/webook/pkg/logger"
)

var serviceProviderSet = wire.NewSet(
	dao.NewCommentDAO,
	repository.NewCommentRepo,
	service.NewCommentSvc,
	grpc2.NewGrpcServer,
)

var thirdProvider = wire.NewSet(
	logger.NewNoOpLogger,
	InitTestDB,
)

func InitGRPCServer() *grpc2.CommentServiceServer {
	wire.Build(thirdProvider, serviceProviderSet)
	return new(grpc2.CommentServiceServer)
}
