package repository

import (
	"context"
	"mbook/webook/internal/repository/cache"
)

var (
	ErrCodeVerifyTooMany = cache.ErrCodeVerifyTooMany
	ErrCodeSendTooMany   = cache.ErrCodeSendTooMany
)

//go:generate mockgen -source=./code.go -package=repomocks -destination=./mocks/code.mock.go CodeRepository
type CodeRepository interface {
	Set(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, code string) (bool, error)
}
type CachedCodeRepository struct {
	cache cache.CodeCache
}

func NewCodeRepository(cache cache.CodeCache) CodeRepository {
	return &CachedCodeRepository{cache: cache}
}

func (repo *CachedCodeRepository) Set(ctx context.Context, biz, phone, code string) error {
	return repo.cache.Set(ctx, biz, phone, code)
}
func (repo *CachedCodeRepository) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	return repo.cache.Verify(ctx, biz, phone, code)
}
