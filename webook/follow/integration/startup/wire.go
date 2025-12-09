//go:build wireinject

package startup

import (
	"github.com/google/wire"
	"mbook/webook/follow/grpc"
	"mbook/webook/follow/repository"
	"mbook/webook/follow/repository/cache"
	"mbook/webook/follow/repository/dao"
	"mbook/webook/follow/service"
)

func InitServer() *grpc.FollowServiceServer {
	wire.Build(
		InitRedis,
		InitLog,
		InitTestDB,
		dao.NewGORMFollowRelationDAO,
		cache.NewRedisFollowCache,
		repository.NewFollowRelationRepository,
		service.NewFollowRelationService,
		grpc.NewFollowRelationServiceServer,
	)
	return new(grpc.FollowServiceServer)
}
