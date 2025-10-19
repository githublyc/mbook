//go:build wireinject

package startup

import (
	"github.com/google/wire"
	"mbook/webook/interactive/grpc"
	repository2 "mbook/webook/interactive/repository"
	cache2 "mbook/webook/interactive/repository/cache"
	dao2 "mbook/webook/interactive/repository/dao"
	service2 "mbook/webook/interactive/service"
)

var thirdPartySet = wire.NewSet(
	//第三方依赖
	InitRedis, InitDB,
	InitLogger,
	//InitSaramaClient,
	//InitSyncProducer,
)
var interactiveSvcSet = wire.NewSet(
	dao2.NewGORMInteractiveDAO,
	cache2.NewInteractiveRedisCache,
	repository2.NewCachedInteractiveRepository,
	service2.NewInteractiveService,
)

func InitInteractiveService() *grpc.InteractiveServiceServer {
	wire.Build(thirdPartySet, interactiveSvcSet, grpc.NewInteractiveServiceServer)
	return new(grpc.InteractiveServiceServer)
}
