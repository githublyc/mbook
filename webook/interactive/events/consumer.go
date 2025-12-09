package events

import (
	"context"
	"github.com/IBM/sarama"
	"mbook/webook/interactive/repository"
	"mbook/webook/pkg/logger"
	"mbook/webook/pkg/saramax"
	"time"
)

const TopicReadEvent = "article_read"

type ReadEvent struct {
	Aid int64
	Uid int64
}
type InteractiveReadEventConsumer struct {
	repo   repository.InteractiveRepository
	client sarama.Client
	l      logger.LoggerV1
}

func NewInteractiveReadEventConsumer(repo repository.InteractiveRepository,
	client sarama.Client, l logger.LoggerV1) *InteractiveReadEventConsumer {
	return &InteractiveReadEventConsumer{repo: repo, client: client, l: l}
}

func (i *InteractiveReadEventConsumer) StartV1() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", i.client)
	if err != nil {
		return err
	}

	go func() {
		er := cg.Consume(context.Background(),
			[]string{TopicReadEvent},
			saramax.NewBatchHandler[ReadEvent](i.l, i.BatchConsume))
		if er != nil {
			i.l.Error("退出消费", logger.Error(er))
		}
	}()
	return err
}

func (i *InteractiveReadEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", i.client)
	if err != nil {
		return err
	}

	go func() {
		// Consume是是一个阻塞调用：一旦开始消费，它会一直在里面跑（拉消息、分发到 handler、处理 rebalance、心跳等）
		// 直到出错或 context 被取消。
		er := cg.Consume(context.Background(),
			[]string{TopicReadEvent},
			saramax.NewHandler[ReadEvent](i.l, i.Consume))
		if er != nil {
			i.l.Error("退出消费", logger.Error(er))
		}
	}()
	return err
}

func (i *InteractiveReadEventConsumer) BatchConsume(msg []*sarama.ConsumerMessage,
	events []ReadEvent) error {
	bizs := make([]string, 0, len(events))
	bizIds := make([]int64, 0, len(events))
	for _, evt := range events {
		bizs = append(bizs, "article")
		bizIds = append(bizIds, evt.Aid)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return i.repo.BatchIncrReadCnt(ctx, bizs, bizIds)
}

func (i *InteractiveReadEventConsumer) Consume(msg *sarama.ConsumerMessage,
	event ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return i.repo.IncrReadCnt(ctx, "article", event.Aid)
}
