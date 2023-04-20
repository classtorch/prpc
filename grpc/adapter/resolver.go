package adapter

import (
	"context"
	pResolver "github.com/classtorch/prpc/resolver"
	"google.golang.org/grpc/attributes"
	gResolver "google.golang.org/grpc/resolver"
)

// resolverBuilder, implementation of gPRC resolver Builder
type resolverBuilder struct {
	notice *Notice
}

// Notice struct used for communication between pPRC and gRPC's Resolver
type Notice struct {
	UpdateState chan pResolver.State //from pRPC resolver UpdateState event, trigger gRPC resolver UpdateState
	Close       chan struct{}        //from gRPC resolver Close event, trigger pRPC resolver Close
	ResolveNow  chan struct{}        //from gRPC resolver ResolveNow event, trigger pRPC resolver ResolveNow
	Ctx         context.Context
	Cancel      context.CancelFunc
}

// new gRPC Resolver Builder
func NewResolverBuilder(notice *Notice) *resolverBuilder {
	return &resolverBuilder{
		notice: notice,
	}
}

// Build method start a goroutine watch pRPC's resolver event
func (tb *resolverBuilder) Build(target gResolver.Target, cc gResolver.ClientConn, opts gResolver.BuildOptions) (gResolver.Resolver, error) {
	go tb.watchPRpcResolver(cc)
	return &resolver{
		tb: tb,
	}, nil
}

// watch pRPC resolver UpdateState change event,then invoke gRPC resolver UpdateState method
func (tb *resolverBuilder) watchPRpcResolver(cc gResolver.ClientConn) {
	for {
		select {
		case state := <-tb.notice.UpdateState:
			cc.UpdateState(buildGrpcState(state))
			break
		case _ = <-tb.notice.Ctx.Done():
			return
		}
	}
}

func (tb *resolverBuilder) Scheme() string {
	return "adaptresolver"
}

// convert pPRC State to gRPC State
func buildGrpcState(stat pResolver.State) gResolver.State {
	grpcAddresses := make([]gResolver.Address, len(stat.Addresses))
	attributesInfo := &attributes.Attributes{}
	for idx, address := range stat.Addresses {
		for key, value := range address.Attributes {
			attributesInfo = attributesInfo.WithValue(key, value)
		}
		grpcAddresses[idx] = gResolver.Address{Addr: address.Addr, Attributes: attributesInfo}
	}
	return gResolver.State{
		Addresses:  grpcAddresses,
		Attributes: attributesInfo,
	}
}

// resolver, implementation of gPRC resolver Builder
type resolver struct {
	tb *resolverBuilder
}

// gRPC Resolver ResolveNow event notice pRPC
func (tr *resolver) ResolveNow(gResolver.ResolveNowOptions) {
	tr.tb.notice.ResolveNow <- struct{}{}
}

// gRPC Resolver Close event notice pRPC
func (tr *resolver) Close() {
	tr.tb.notice.Close <- struct{}{}
}
