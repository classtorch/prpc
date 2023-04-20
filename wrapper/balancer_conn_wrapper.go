package wrapper

import (
	"github.com/classtorch/prpc/balancer"
	"github.com/classtorch/prpc/resolver"
	"sync"
)

// CCBalancerWrapper
type CCBalancerWrapper struct {
	pickWrapper *PickerWrapper
	balancer    balancer.Balancer
	mu          sync.Mutex
}

// NewCCBalancerWrapper return a CCBalancerWrapper
func NewCCBalancerWrapper(pickWrapper *PickerWrapper, b balancer.Builder) *CCBalancerWrapper {
	ccb := &CCBalancerWrapper{
		pickWrapper: pickWrapper,
	}
	ccb.balancer = b.Build()
	return ccb
}

func (ccb *CCBalancerWrapper) close() {
}

// UpdateState build a picker
func (ccb *CCBalancerWrapper) UpdateState(state resolver.State) {
	ccb.mu.Lock()
	defer ccb.mu.Unlock()
	newPicker, _ := ccb.balancer.UpdateState(state)
	ccb.pickWrapper.updatePicker(newPicker)
}
