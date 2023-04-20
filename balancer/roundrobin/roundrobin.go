package roundrobin

import (
	"github.com/classtorch/prpc/balancer"
	"github.com/classtorch/prpc/balancer/base"
	"github.com/classtorch/prpc/resolver"
	"sync"
)

var (
	Name = "round_robin"
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
	return &rbPicker{next: 0, addrs: info.ReadyAddresses}
}

type rbPicker struct {
	mu    sync.Mutex
	addrs []resolver.Address
	next  int
}

func (rp *rbPicker) Pick() (balancer.PickResult, error) {
	rp.mu.Lock()
	defer rp.mu.Unlock()
	if len(rp.addrs) == 0 {
		return balancer.PickResult{}, balancer.ErrNoAddressAvailable
	}
	addr := rp.addrs[rp.next]
	rp.next = (rp.next + 1) % len(rp.addrs)
	return balancer.PickResult{Address: addr}, nil
}
