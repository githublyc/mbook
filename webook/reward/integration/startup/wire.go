//go:build wireinject

package startup

import (
	"github.com/google/wire"
	pmtv1 "mbook/webook/api/proto/gen/payment/v1"
	"mbook/webook/reward/repository"
	"mbook/webook/reward/repository/cache"
	"mbook/webook/reward/repository/dao"
	"mbook/webook/reward/service"
)

var thirdPartySet = wire.NewSet(InitTestDB, InitLogger, InitRedis)

func InitWechatNativeSvc(client pmtv1.WechatPaymentServiceClient) *service.WechatNativeRewardService {
	wire.Build(service.NewWechatNativeRewardService,
		thirdPartySet,
		cache.NewRewardRedisCache,
		repository.NewRewardRepository, dao.NewRewardGORMDAO)
	return new(service.WechatNativeRewardService)
}
