package weight

import (
	"github.com/classtorch/prpc/balancer"
	"github.com/classtorch/prpc/balancer/base"
	"github.com/classtorch/prpc/resolver"
	"sync"
)

var (
	Name = "weight_round_robin"
)

func NewBalancerBuild() balancer.Builder {
	return base.NewBalancerBuilder(Name, &rbPickerBuilder{})
}

func init() {
	balancer.Register(NewBalancerBuild())
}

type rbPickerBuilder struct {
}

func (rpb *rbPickerBuilder) Build(info balancer.PickerBuildInfo) balancer.Picker {
	readyAddressesMap := make(map[string]resolver.Address, 0)
	weightRoundRobin := &RoundRobin{}
	for _, addr := range info.ReadyAddresses {
		readyAddressesMap[addr.Addr] = addr
		if weight, ok := addr.Attributes["weight"].(int); ok {
			weightRoundRobin.Add(addr.Addr, weight)
		}
	}
	return &rbPicker{readyAddressesMap: readyAddressesMap, weightRoundRobin: weightRoundRobin}
}

type rbPicker struct {
	readyAddressesMap map[string]resolver.Address
	mu                sync.Mutex
	weightRoundRobin  *RoundRobin
}

//Pick select next address
func (rp *rbPicker) Pick() (balancer.PickResult, error) {
	rp.mu.Lock()
	defer rp.mu.Unlock()
	selectAddr := rp.weightRoundRobin.Next()
	if len(selectAddr) == 0 {
		return balancer.PickResult{}, balancer.ErrNoAddressAvailable
	}
	addrResult, ok := rp.readyAddressesMap[selectAddr]
	if !ok {
		addrResult = resolver.Address{Addr: selectAddr}
	}
	return balancer.PickResult{Address: addrResult}, nil
}

//RoundRobin a weight roundRobin container
type RoundRobin struct {
	curIndex int
	nodes    []*Node
}

//Node a weight node
type Node struct {
	addr            string
	weight          int
	currentWeight   int
	effectiveWeight int
}

//Add add a node
func (r *RoundRobin) Add(addr string, weight int) {
	r.nodes = append(r.nodes, &Node{addr: addr, weight: weight, effectiveWeight: weight})
}

//Next get a next address
func (r *RoundRobin) Next() string {
	total := 0
	var best *Node
	for idx, node := range r.nodes {
		total += node.effectiveWeight
		node.currentWeight += node.effectiveWeight
		if node.effectiveWeight < node.weight {
			node.effectiveWeight++
		}
		if best == nil || node.currentWeight > best.currentWeight {
			best = node
			r.curIndex = idx
		}
	}
	if best == nil {
		return ""
	}
	best.currentWeight -= total
	return best.addr
}
