package wrr

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"sync"
)

const Name = "custom_weighted_round_robin"

func newBuilder() balancer.Builder {
	return base.NewBalancerBuilder(Name, &PickerBuilder{},
		base.Config{HealthCheck: true})
}

func init() {
	balancer.Register(newBuilder())
}

type PickerBuilder struct{}

func (p *PickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]*weightedConn, 0, len(info.ReadySCs))
	for sc, sci := range info.ReadySCs {
		md, _ := sci.Address.Metadata.(map[string]any)
		weightVal, _ := md["weight"]
		weight, _ := weightVal.(float64)
		//if weight==0 {
		//	//如果上面3个不ok，可以给个默认值
		//}
		conns = append(conns, &weightedConn{
			SubConn:       sc,
			weight:        int(weight),
			currentWeight: int(weight),
		})
	}
	return &Picker{
		conns: conns,
	}
}

type Picker struct {
	conns []*weightedConn
	lock  sync.Mutex
}

func (p *Picker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if len(p.conns) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	// 总权重
	var total int
	var maxCC *weightedConn
	for _, conn := range p.conns {
		total += conn.weight
		conn.currentWeight += conn.weight
		if maxCC == nil || conn.currentWeight > maxCC.currentWeight {
			maxCC = conn
		}
	}

	maxCC.currentWeight -= total

	return balancer.PickResult{
		SubConn: maxCC.SubConn,
		Done: func(info balancer.DoneInfo) {
			// failover 要在这里做文章
			// 根据调用结果的具体错误信息进行容错
			// 1. 如果要是触发了限流了，
			// 1.1 你可以考虑直接挪走这个节点，后面再挪回来
			// 1.2 你可以考虑直接将 weight/currentWeight 调整到极低
			// 2. 触发了熔断呢？
			// 3. 降级呢？
		},
	}, nil
}

// 组合了SubConn
type weightedConn struct {
	balancer.SubConn
	// 初始权重
	weight        int
	currentWeight int
}
