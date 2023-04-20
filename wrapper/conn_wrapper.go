package wrapper

import (
	"github.com/classtorch/prpc/balancer"
	"github.com/classtorch/prpc/balancer/roundrobin"
)

// ConnWrapper
type ConnWrapper struct {
	ResolverWrapper *CCResolverWrapper
	BalancerWrapper *CCBalancerWrapper
	PickerWrapper   *PickerWrapper
}

func GetBalancerWrapper(curBalancerName string, pickerWrapper *PickerWrapper) (*CCBalancerWrapper, error) {
	var balancerBuilder balancer.Builder
	if len(curBalancerName) > 0 {
		balancerBuilder = balancer.Get(curBalancerName)
		if balancerBuilder == nil {
			return nil, balancer.BalancerNotExistErr
		}
	}
	if balancerBuilder == nil {
		balancerBuilder = balancer.Get(roundrobin.Name)
	}
	return NewCCBalancerWrapper(pickerWrapper, balancerBuilder), nil
}
