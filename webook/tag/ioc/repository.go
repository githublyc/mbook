package ioc

import (
	"context"
	"mbook/webook/pkg/logger"
	"mbook/webook/tag/repository"
	"mbook/webook/tag/repository/cache"
	"mbook/webook/tag/repository/dao"
	"time"
)

func InitRepository(d dao.TagDAO, c cache.TagCache, l logger.LoggerV1) repository.TagRepository {
	repo := repository.NewTagRepository(d, c, l)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		// 也可以同步执行。但是在一些场景下，同步执行会占用很长的时间，所以可以考虑异步执行。
		repo.PreloadUserTags(ctx)
	}()
	return repo
}
