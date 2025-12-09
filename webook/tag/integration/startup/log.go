package startup

import (
	"mbook/webook/pkg/logger"
)

func InitLog() logger.LoggerV1 {
	return logger.NewNoOpLogger()
}
