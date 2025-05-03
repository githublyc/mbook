package startup

import (
	"mbook/webook/internal/service/oauth2/wechat"
	"mbook/webook/pkg/logger"
)

func InitWechatService(l logger.LoggerV1) wechat.Service {
	return wechat.NewService("appID", "appSecret", l)
}
