package startup

import (
	"mbook/webook/pkg/logger"
	"mbook/webook/tag/repository"
	"mbook/webook/tag/repository/cache"
	"mbook/webook/tag/repository/dao"
)

func InitRepository(d dao.TagDAO, c cache.TagCache, l logger.LoggerV1) repository.TagRepository {
	return repository.NewTagRepository(d, c, l)
}
