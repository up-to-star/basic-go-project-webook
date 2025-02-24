package wrr

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"sync"
)

const name = "custom_wrr"

func init() {
	// 把一个pickerBuilder 转换成一个 builder
	balancer.Register(base.NewBalancerBuilder(name, &PickerBuilder{}, base.Config{}))
}

type PickerBuilder struct {
}

func (p *PickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]*conn, 0, len(info.ReadySCs))
	for subConn, subConnInfo := range info.ReadySCs {
		cc := &conn{
			cc: subConn,
		}
		md, ok := subConnInfo.Address.Metadata.(map[string]any)
		if ok {
			weightVal := md["weight"]
			weight, _ := weightVal.(float64)
			cc.weight = int(weight)
		}
		if cc.weight == 0 {
			// 默认权重
			cc.weight = 10
		}
		cc.currentWeight = cc.weight
		conns = append(conns, cc)
	}
	return &Picker{}
}

// Picker 真正实现负载均衡的地方
type Picker struct {
	conns []*conn
	mutex sync.Mutex
}

// Pick 实现基于权重的负载均衡算法
func (p *Picker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(p.conns) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	var total int
	p.mutex.Lock()
	mxCC := p.conns[0]
	for _, cc := range p.conns {
		total += cc.weight
		cc.currentWeight += cc.weight
		if cc.currentWeight > mxCC.currentWeight {
			mxCC = cc
		}
	}
	mxCC.currentWeight -= total
	p.mutex.Unlock()
	return balancer.PickResult{
		SubConn: mxCC.cc,
		Done: func(info balancer.DoneInfo) {
			// 可以根据调用结果调整权重
		},
	}, nil
}

// conn 代表节点
type conn struct {
	weight        int
	currentWeight int
	// grpc 里面代表一个节点的表达
	cc balancer.SubConn
}
