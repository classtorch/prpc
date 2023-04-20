package base

import (
	"github.com/classtorch/prpc/balancer"
	"github.com/classtorch/prpc/resolver"
)

type baseBuilder struct {
	name          string
	pickerBuilder balancer.PickerBuilder
}

func (bb *baseBuilder) Build() balancer.Balancer {
	bal := &baseBalancer{
		pickerBuilder: bb.pickerBuilder,
	}
	return bal
}

func (bb *baseBuilder) Name() string {
	return bb.name
}

type baseBalancer struct {
	pickerBuilder balancer.PickerBuilder
	picker        balancer.Picker
}

func (b *baseBalancer) UpdateState(state resolver.State) (balancer.Picker, error) {
	b.picker = b.pickerBuilder.Build(balancer.PickerBuildInfo{ReadyAddresses: state.Addresses})
	return b.picker, nil
}

func (b *baseBalancer) Close() {

}
