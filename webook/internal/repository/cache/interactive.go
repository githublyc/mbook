package cache

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/redis/go-redis/v9"
)

var (
	//go:embed lua/incr_cnt.lua
	luaInrcCnt string
)

const fieldReadCnt = "read_cnt"
const fieldLikeCnt = "like_cnt"
const fieldCollectCnt = "collect_cnt"

type InteractiveCache interface {
	IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error
	IncrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error
	DecrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error
	IncrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error
}
type InteractiveRedisCache struct {
	client redis.Cmdable
}

func (i *InteractiveRedisCache) IncrCollectCntIfPresent(ctx context.Context,
	biz string, bizId int64) error {
	key := i.key(biz, bizId)
	// 不是特别需要处理 res , res=0可以接受
	//_, err := i.client.Eval(ctx, luaInrcCnt,
	//	[]string{key}, fieldReadCnt, 1).Int()
	//return err
	return i.client.Eval(ctx, luaInrcCnt,
		[]string{key}, fieldCollectCnt, 1).Err()
}

func (i *InteractiveRedisCache) IncrLikeCntIfPresent(ctx context.Context,
	biz string, bizId int64) error {
	key := i.key(biz, bizId)
	// 不是特别需要处理 res , res=0可以接受
	//_, err := i.client.Eval(ctx, luaInrcCnt,
	//	[]string{key}, fieldReadCnt, 1).Int()
	//return err
	return i.client.Eval(ctx, luaInrcCnt,
		[]string{key}, fieldLikeCnt, 1).Err()
}

func (i *InteractiveRedisCache) DecrLikeCntIfPresent(ctx context.Context,
	biz string, bizId int64) error {
	key := i.key(biz, bizId)
	// 不是特别需要处理 res , res=0可以接受
	//_, err := i.client.Eval(ctx, luaInrcCnt,
	//	[]string{key}, fieldReadCnt, 1).Int()
	//return err
	return i.client.Eval(ctx, luaInrcCnt,
		[]string{key}, fieldLikeCnt, -1).Err()
}

func NewInteractiveRedisCache(client redis.Cmdable) InteractiveCache {
	return &InteractiveRedisCache{client: client}
}

func (i *InteractiveRedisCache) IncrReadCntIfPresent(ctx context.Context,
	biz string, bizId int64) error {
	key := i.key(biz, bizId)
	// 不是特别需要处理 res , res=0可以接受
	//_, err := i.client.Eval(ctx, luaInrcCnt,
	//	[]string{key}, fieldReadCnt, 1).Int()
	//return err
	return i.client.Eval(ctx, luaInrcCnt,
		[]string{key}, fieldReadCnt, 1).Err()

}
func (i *InteractiveRedisCache) key(biz string, bizId int64) string {
	return fmt.Sprintf("interactive:%s:%d", biz, bizId)
}
