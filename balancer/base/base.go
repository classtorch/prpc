package base

import (
	"github.com/classtorch/prpc/balancer"
)

func NewBalancerBuilder(name string, pb balancer.PickerBuilder) balancer.Builder {
	return &baseBuilder{
		name:          name,
		pickerBuilder: pb,
	}
}
