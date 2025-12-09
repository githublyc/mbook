package job

import (
	"context"
	"mbook/webook/payment/service/wechat"
	"mbook/webook/pkg/logger"
	"time"
)

type SyncWechatOrderJob struct {
	svc *wechat.NativePaymentService
	l   logger.LoggerV1
}

func (s *SyncWechatOrderJob) Name() string {
	return "sync_wechat_order_job"
}

// 不必特别频繁，比如说一分钟运行一次
func (s *SyncWechatOrderJob) Run() error {
	// 定时找到超时的微信支付订单，然后发起同步
	// 针对过期订单
	t := time.Now().Add(-time.Minute * 31)
	offset := 0
	const limit = 100
	for {
		ctx, cancel := context.WithTimeout(context.Background(),
			time.Second*3)
		pmts, err := s.svc.FindExpiredPayment(ctx, offset, limit, t)
		cancel()
		if err != nil {
			return err
		}
		// SyncWechatInfo不是批量的接口，所以只能一个一个的处理了
		for _, pmt := range pmts {
			ctx, cancel = context.WithTimeout(context.Background(),
				time.Second*3)
			err = s.svc.SyncWechatInfo(ctx, pmt.BizTradeNO)
			cancel()
			if err != nil {
				s.l.Error("同步微信订单状态失败", logger.Error(err),
					logger.String("biz_trade_no", pmt.BizTradeNO))
			}
		}
		if len(pmts) < limit {
			return nil
		}
		offset += limit
	}
}
