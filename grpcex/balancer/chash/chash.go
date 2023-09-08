package chash

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"log"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

type ChashCtxKey string

const (
	Name                    = "chash"
	chashCtxKey ChashCtxKey = "chash_key"
)

var (
	chashKeyTypeUnknow = errors.New("chash key type unknown")
)

func init() {
	builder := base.NewBalancerBuilder(Name, &chashPickerBuilder{}, base.Config{HealthCheck: true})
	balancer.Register(builder)
}

type chashPickerBuilder struct{}

func (*chashPickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	log.Printf("Build chash balancer.Picker: %v\n", info)
	if len(info.ReadySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}

	p := &chashPicker{}
	for sc, info := range info.ReadySCs {
		for i := 0; i < 4; i++ {
			addr := fmt.Sprintf("%d/%s", i, info.Address.Addr)
			p.AddSubConn(addr, sc)
		}
	}
	return p
}

type chashPicker struct {
	nodes []*chashNode
}

type chashNode struct {
	hash    uint32
	subConn balancer.SubConn
}

func (p *chashPicker) AddNode(node *chashNode) {
	for i, n := range p.nodes {
		if n.hash > node.hash {
			p.nodes = append(p.nodes[:i], append([]*chashNode{node}, p.nodes[i:]...)...)
			return
		}
	}
	p.nodes = append(p.nodes, node)
}

func (p *chashPicker) AddSubConn(addr string, subConn balancer.SubConn) {
	hash := bkdrHash32([]byte(addr))
	node := &chashNode{
		hash:    hash,
		subConn: subConn,
	}
	p.AddNode(node)
}

func (p *chashPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	key := info.Ctx.Value(chashCtxKey)
	var hash uint32
	switch value := key.(type) {
	case int64:
		hash = bkdrHash32Int64(value)
	case string:
		hash = bkdrHash32([]byte(value))
	default:
		return balancer.PickResult{}, chashKeyTypeUnknow
	}

	for _, n := range p.nodes {
		if n.hash > hash {
			return balancer.PickResult{SubConn: n.subConn}, nil
		}
	}
	return balancer.PickResult{SubConn: p.nodes[0].subConn}, nil
}

func WithChashKey(ctx context.Context, key any) context.Context {
	return context.WithValue(ctx, chashCtxKey, key)
}

func bkdrHash32(data []byte) uint32 {
	var hash uint32 = 0
	for _, b := range data {
		hash = hash*31 + uint32(b)
	}
	return hash
}

func bkdrHash32Int64(val int64) uint32 {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, uint64(val))
	return bkdrHash32(data)
}
