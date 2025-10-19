package cache

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/redis/go-redis/v9"
	"mbook/webook/interactive/domain"
	"strconv"
	"time"
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
	Get(ctx context.Context, biz string, id int64) (domain.Interactive, error)
	Set(ctx context.Context, biz string, id int64, intr domain.Interactive) error
}
type InteractiveRedisCache struct {
	client redis.Cmdable
}

func (i *InteractiveRedisCache) Get(ctx context.Context,
	biz string, id int64) (domain.Interactive, error) {
	key := i.key(biz, id)
	res, err := i.client.HGetAll(ctx, key).Result()
	if err != nil {
		return domain.Interactive{}, err
	}
	if len(res) == 0 {
		return domain.Interactive{}, ErrKeyNotExist
	}
	var intr domain.Interactive
	intr.BizId = id
	//可以忽略错误
	intr.ReadCnt, _ = strconv.ParseInt(res[fieldReadCnt], 10, 64)
	intr.LikeCnt, _ = strconv.ParseInt(res[fieldLikeCnt], 10, 64)
	intr.CollectCnt, _ = strconv.ParseInt(res[fieldCollectCnt], 10, 64)
	return intr, nil
}

func (i *InteractiveRedisCache) Set(ctx context.Context,
	biz string, id int64, intr domain.Interactive) error {
	key := i.key(biz, id)
	err := i.client.HSet(ctx, key, fieldReadCnt, intr.ReadCnt,
		fieldLikeCnt, intr.LikeCnt,
		fieldCollectCnt, intr.CollectCnt).Err()
	if err != nil {
		return err
	}
	return i.client.Expire(ctx, key, time.Minute*15).Err()
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
