//go:build wireinject

package startup

import (
	"github.com/google/wire"
	"mbook/webook/tag/events"
	"mbook/webook/tag/grpc"
	"mbook/webook/tag/repository/cache"
	"mbook/webook/tag/repository/dao"
	"mbook/webook/tag/service"
)

func InitGRPCService(p events.Producer) *grpc.TagServiceServer {
	wire.Build(InitTestDB, InitRedis,
		InitLog,
		dao.NewGORMTagDAO,
		InitRepository,
		cache.NewRedisTagCache,
		service.NewTagService,
		grpc.NewTagServiceServer,
	)
	return new(grpc.TagServiceServer)
}
