package repository

import (
	"context"
	"mbook/webook/internal/domain"
	"mbook/webook/internal/repository/cache"
)

type RankingRepository interface {
	ReplaceTopN(ctx context.Context, arts []domain.Article) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
}
type CachedOnlyRankingRepository struct {
	cache cache.RankingCache
	// 下面是给 v1 用的
	redisCache *cache.RankingRedisCache
	localCache *cache.RankingLocalCache
}

func NewCachedOnlyRankingRepositoryV1(redisCache *cache.RankingRedisCache,
	localCache *cache.RankingLocalCache) *CachedOnlyRankingRepository {
	return &CachedOnlyRankingRepository{redisCache: redisCache, localCache: localCache}
}

func (repo *CachedOnlyRankingRepository) GetTopN(ctx context.Context) ([]domain.Article, error) {
	return repo.cache.Get(ctx)
}
func (repo *CachedOnlyRankingRepository) GetTopNV1(ctx context.Context) ([]domain.Article, error) {
	res, err := repo.localCache.Get(ctx)
	if err == nil {
		return res, nil
	}
	res, err = repo.redisCache.Get(ctx)
	if err != nil {
		//Redis崩溃的时候，再次尝试从本地缓存获取(此时不去检查本地缓存是否过期)
		return repo.localCache.ForceGet(ctx)
	}
	_ = repo.localCache.Set(ctx, res)
	return res, nil
}

func NewCachedOnlyRankingRepository(cache cache.RankingCache) RankingRepository {
	return &CachedOnlyRankingRepository{cache: cache}
}

func (repo *CachedOnlyRankingRepository) ReplaceTopN(ctx context.Context,
	arts []domain.Article) error {
	return repo.cache.Set(ctx, arts)
}
func (repo *CachedOnlyRankingRepository) ReplaceTopNV1(ctx context.Context,
	arts []domain.Article) error {
	_ = repo.localCache.Set(ctx, arts)
	return repo.redisCache.Set(ctx, arts)
}
