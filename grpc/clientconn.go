package grpc

import (
	"context"
	"fmt"
	"github.com/classtorch/prpc/balancer"
	"github.com/classtorch/prpc/balancer/roundrobin"
	"github.com/classtorch/prpc/grpc/adapter"
	"github.com/classtorch/prpc/logger"
	"github.com/classtorch/prpc/resolver"
	"github.com/classtorch/prpc/wrapper"
	"google.golang.org/grpc"
	"log"
	"os"
	"sync"
)

// ClientConn pRPC's grpc ClientConn
type ClientConn struct {
	connOption      connectOption
	mu              sync.Mutex
	Target          string
	parseTarget     resolver.Target //parse url from consul
	resolverWrapper *wrapper.CCResolverWrapper
	balancerWrapper *wrapper.CCBalancerWrapper
	pickerWrapper   *wrapper.PickerWrapper
	notice          *adapter.Notice
}

// connectOption pRPC's grpc ClientConn connect Option
type connectOption struct {
	grpcOpts      []grpc.DialOption
	resolvers     []resolver.Builder
	balancerBuild balancer.Builder
	// selected balancer name
	curBalancerName string
	log             logger.Log
}

type CallOption struct {
	grpc.CallOption
}

func defaultConnectOption() connectOption {
	return connectOption{
		curBalancerName: roundrobin.Name,
		log:             logger.NewDefaultLogger(log.New(os.Stderr, logger.DefaultLogPrefix, log.LstdFlags)),
	}
}

type ConnOption func(*connectOption)

func WithResolver(resolvers ...resolver.Builder) ConnOption {
	return func(o *connectOption) {
		o.resolvers = append(o.resolvers, resolvers...)
	}
}

func WithBalancerName(name string) ConnOption {
	return func(o *connectOption) {
		o.curBalancerName = name
	}
}

func WithOptions(grpcOpts ...grpc.DialOption) ConnOption {
	return func(o *connectOption) {
		o.grpcOpts = grpcOpts
	}
}

func WithLog(log logger.Log) ConnOption {
	return func(o *connectOption) {
		o.log = log
	}
}

// NewClientConn  init a pRPC's grpc ClientConn, and build communication between pPRC and gRPC's gResolver
func NewClientConn(ctx context.Context, target string, opts ...ConnOption) (*grpc.ClientConn, error) {
	cc := &ClientConn{
		Target:     target,
		connOption: defaultConnectOption(),
	}
	for _, option := range opts {
		option(&cc.connOption)
	}
	parseTarget := resolver.ParseTarget(target)
	scheme := parseTarget.Scheme
	if scheme == resolver.GetPassThroughScheme() {
		return grpc.DialContext(ctx, target, cc.connOption.grpcOpts...)
	}
	cc.parseTarget = parseTarget

	pickerWrapper := wrapper.NewPickerWrapper(cc.connOption.log)
	cc.pickerWrapper = pickerWrapper
	balancerWrapper, err := wrapper.GetBalancerWrapper(cc.connOption.curBalancerName, pickerWrapper)
	if err != nil {
		return nil, err
	}
	cc.balancerWrapper = balancerWrapper

	resolverBuild := cc.getResolverBuilder(cc.parseTarget.Scheme)
	if resolverBuild == nil {
		return nil, resolver.ResolverNotExistErr
	}

	noticeCtx, cancel := context.WithCancel(context.Background())
	cc.notice = &adapter.Notice{
		UpdateState: make(chan resolver.State),
		Close:       make(chan struct{}),
		ResolveNow:  make(chan struct{}),
		Ctx:         noticeCtx,
		Cancel:      cancel,
	}
	go cc.watchGrpcResolver()

	resolverWrapper, err := wrapper.NewCCResolverWrapper(cc, resolverBuild)
	if err != nil {
		return nil, err
	}
	cc.resolverWrapper = resolverWrapper

	grpcClientConn, err := newGrpcClientConn(ctx, cc.notice, pickerWrapper, parseTarget, cc.connOption.grpcOpts)
	if err != nil {
		// if happen err,need to cancel notice,let watch goroutine exit,this is very important
		cc.notice.Cancel()
		return nil, err
	}
	return grpcClientConn, nil
}

// newGrpcClientConn new grpc ClientConn
func newGrpcClientConn(ctx context.Context, notice *adapter.Notice, pickerWrapper *wrapper.PickerWrapper, target resolver.Target, options []grpc.DialOption) (*grpc.ClientConn, error) {
	adaptResolverBuilder := adapter.NewResolverBuilder(notice)
	adaptBalancerName := adapter.RegisterBalancer(pickerWrapper)
	grpcOpts := options
	grpcOpts = append(grpcOpts, grpc.WithResolvers(adaptResolverBuilder),
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingPolicy":"%s"}`, adaptBalancerName)))
	newTarget := fmt.Sprintf("%s://%s/%s", adaptResolverBuilder.Scheme(), target.Agent, target.Endpoint)
	return grpc.DialContext(
		ctx,
		newTarget,
		grpcOpts...,
	)
}

// watchGrpcResolver watch grpc resolver Close and ResolveNow event
// when receive event, call pRPC's resolver Close and ResolveNow method
func (cc *ClientConn) watchGrpcResolver() {
	for {
		select {
		case _ = <-cc.notice.Close:
			cc.resolverWrapper.Close()
		case _ = <-cc.notice.ResolveNow:
			cc.resolverWrapper.ResolveNow()
		case _ = <-cc.notice.Ctx.Done():
			return
		}
	}
}

// getResolverBuilder return resolver Builder
func (cc *ClientConn) getResolverBuilder(scheme string) resolver.Builder {
	for _, rb := range cc.connOption.resolvers {
		if scheme == rb.Scheme() {
			return rb
		}
	}
	return resolver.Get(scheme)
}

// UpdateResolverState call pPRC balancer UpdateState and notice gRPC
// please note that notifying gRPC updateState opens a goroutine, and it will block if the goroutine is not used,
// because UpdateState channel not ready to receive，more specifically, the adapter watchPRpcResolver method is not executed
func (cc *ClientConn) UpdateResolverState(state resolver.State, err error) error {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.balancerWrapper.UpdateState(state)
	go func() {
		cc.notice.UpdateState <- state
	}()
	return nil
}

// GetParsedTarget return parsed target
func (cc *ClientConn) GetParsedTarget() resolver.Target {
	return cc.parseTarget
}
