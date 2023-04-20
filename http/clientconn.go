package http

import (
	"context"
	"github.com/classtorch/prpc/balancer"
	"github.com/classtorch/prpc/logger"
	"github.com/classtorch/prpc/resolver"
	"github.com/classtorch/prpc/wrapper"
	"log"
	"os"
	"sync"
	"time"
)

const (
	defaultTimeOut = time.Second * 5 // 全局http超时时间
	defaultSecure  = false
)

// ClientConn http ClientConn
type ClientConn struct {
	ctx             context.Context
	target          string
	direct          bool
	connOption      connectOption
	mu              sync.Mutex
	parseTarget     resolver.Target //parse url from consul
	resolverWrapper *wrapper.CCResolverWrapper
	balancerWrapper *wrapper.CCBalancerWrapper
	pickerWrapper   *wrapper.PickerWrapper
}

// connectOption http ClientConn connect Option
type connectOption struct {
	timeOut          time.Duration
	resolverBuilder  []resolver.Builder
	balancerBuilder  balancer.Builder
	chanInterceptor  []Interceptor
	unaryInterceptor Interceptor
	curBalancerName  string
	httpCall         CallInterface
	secure           bool
	log              logger.Log
}

func defaultConnectOption() connectOption {
	return connectOption{
		timeOut: defaultTimeOut,
		secure:  defaultSecure,
		log:     logger.NewDefaultLogger(log.New(os.Stderr, logger.DefaultLogPrefix, log.LstdFlags)),
	}
}

type ConnOption func(*connectOption)

func WithTimeOut(value int) ConnOption {
	return func(o *connectOption) {
		o.timeOut = time.Second * time.Duration(value)
	}
}

func WithResolver(resolvers ...resolver.Builder) ConnOption {
	return func(o *connectOption) {
		o.resolverBuilder = append(o.resolverBuilder, resolvers...)
	}
}

func WithSecure(secure bool) ConnOption {
	return func(o *connectOption) {
		o.secure = secure
	}
}

func WithCallClient(httpCall CallInterface) ConnOption {
	return func(o *connectOption) {
		o.httpCall = httpCall
	}
}

func WithBalancerName(name string) ConnOption {
	return func(o *connectOption) {
		o.curBalancerName = name
	}
}

func WithInterceptor(interceptors ...Interceptor) ConnOption {
	return func(o *connectOption) {
		o.chanInterceptor = append(o.chanInterceptor, interceptors...)
	}
}

func WithLog(log logger.Log) ConnOption {
	return func(o *connectOption) {
		o.log = log
	}
}

// NewClientConn init a http ClientConn
func NewClientConn(ctx context.Context, target string, opts ...ConnOption) (*ClientConn, error) {
	cc := &ClientConn{
		target:     target,
		connOption: defaultConnectOption(),
	}
	for _, option := range opts {
		option(&cc.connOption)
	}

	if cc.connOption.httpCall == nil {
		cc.connOption.httpCall = NewDefaultPRpcHttpClient()
	}

	ChainUnaryInterceptors(cc)

	parseTarget := resolver.ParseTarget(target)
	scheme := parseTarget.Scheme
	if scheme == resolver.GetPassThroughScheme() {
		cc.direct = true
		return cc, nil
	}
	cc.parseTarget = parseTarget

	pickerWrapper := wrapper.NewPickerWrapper(cc.connOption.log)
	cc.pickerWrapper = pickerWrapper
	balancerWrapper, err := wrapper.GetBalancerWrapper(cc.connOption.curBalancerName, pickerWrapper)
	if err != nil {
		return nil, err
	}
	cc.balancerWrapper = balancerWrapper
	// init resolver
	resolverBuild := cc.getResolverBuilder(cc.parseTarget.Scheme)
	if resolverBuild == nil {
		return nil, resolver.ResolverNotExistErr
	}
	resolverWrapper, err := wrapper.NewCCResolverWrapper(cc, resolverBuild)
	if err != nil {
		return nil, err
	}
	cc.resolverWrapper = resolverWrapper
	return cc, nil
}

// getResolverBuilder return resolver Builder
func (cc *ClientConn) getResolverBuilder(scheme string) resolver.Builder {
	for _, rb := range cc.connOption.resolverBuilder {
		if scheme == rb.Scheme() {
			return rb
		}
	}
	return resolver.Get(scheme)
}

// UpdateResolverState update balancer State
func (cc *ClientConn) UpdateResolverState(state resolver.State, err error) error {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.balancerWrapper.UpdateState(state)
	return nil
}

// GetParsedTarget return parsed target
func (cc *ClientConn) GetParsedTarget() resolver.Target {
	return cc.parseTarget
}

// GetOption return ClientConn connOption
func (cc *ClientConn) GetOption() connectOption {
	return cc.connOption
}

// GetPickerWrapper return PickerWrapper
func (cc *ClientConn) GetPickerWrapper() *wrapper.PickerWrapper {
	return cc.pickerWrapper
}
