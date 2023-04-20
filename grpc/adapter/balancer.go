package adapter

import (
	"errors"
	"github.com/classtorch/prpc/wrapper"
	gBalancer "google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"sync"
)

// adaptGrpcBalancerName is the name of balancer_for_adapt_grpc balancer.
const adaptGrpcBalancerName = "balancer_for_adapt_grpc"

//RegisterBalancer, register a balancer to gRPC
func RegisterBalancer(picker *wrapper.PickerWrapper) string {
	builder := base.NewBalancerBuilder(adaptGrpcBalancerName, &rrPickerBuilder{pickerWrapper: picker}, base.Config{HealthCheck: true})
	gBalancer.Register(builder)
	return adaptGrpcBalancerName
}

// rrPicker,implementation of gPRC base balancer PickerBuilder
type rrPickerBuilder struct {
	pickerWrapper *wrapper.PickerWrapper
}

// gRPC Build, Build a gRPC Picker
func (pb *rrPickerBuilder) Build(info base.PickerBuildInfo) gBalancer.Picker {
	if len(info.ReadySCs) == 0 {
		return base.NewErrPicker(gBalancer.ErrNoSubConnAvailable)
	}
	return &rrPicker{
		pickerWrapper: pb.pickerWrapper,
		readySCs:      info.ReadySCs,
	}
}

// rrPicker,implementation of gPRC balancer Picker
type rrPicker struct {
	// pPRC PickerWrapper
	pickerWrapper *wrapper.PickerWrapper
	// gRPC readySCs
	readySCs map[gBalancer.SubConn]base.SubConnInfo
	mu       sync.Mutex
}

// gRPC Pick, get a available SubConn
// When Pick is called, the Pick method of pRPC will be called
func (p *rrPicker) Pick(pickInfo gBalancer.PickInfo) (gBalancer.PickResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	pickResult, err := p.pickerWrapper.Pick(pickInfo.Ctx, false)
	if err != nil {
		return gBalancer.PickResult{}, err
	}
	for subConn, subInfo := range p.readySCs {
		if pickResult == subInfo.Address.Addr {
			return gBalancer.PickResult{SubConn: subConn}, nil
		}
	}
	return gBalancer.PickResult{SubConn: nil}, errors.New("not found subConn")
}
