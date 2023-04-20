package prpc

import (
	"context"
	"errors"
	"github.com/classtorch/prpc/grpc"
	"github.com/classtorch/prpc/http"
	grpcRaw "google.golang.org/grpc"
)

// ClientConnInterface
type ClientConnInterface interface {
	HttpInvoke(ctx context.Context, method string, api string, req interface{}, reply interface{}, opts ...http.CallOption) error
	GrpcInvoke(ctx context.Context, method string, req interface{}, reply interface{}, opts ...grpc.CallOption) error
	NewStream(ctx context.Context, desc *grpcRaw.StreamDesc, method string, opts ...grpc.CallOption) (grpcRaw.ClientStream, error)
}

// ClientConn is pRPC ClientConn
type ClientConn struct {
	httpConn *http.ClientConn
	grpcConn *grpcRaw.ClientConn
}

func (cc *ClientConn) GetHttpConn() *http.ClientConn {
	return cc.httpConn
}

func (cc *ClientConn) GetGrpcConn() *grpcRaw.ClientConn {
	return cc.grpcConn
}

func NewClientConn() *ClientConn {
	return &ClientConn{}
}

func (cc *ClientConn) NewGrpcClientConn(ctx context.Context, target string, opts ...grpc.ConnOption) error {
	grpcConn, err := grpc.NewClientConn(ctx, target, opts...)
	if err != nil {
		return err
	}
	cc.grpcConn = grpcConn
	return nil
}

func (cc *ClientConn) NewHttpClientConn(ctx context.Context, target string, opts ...http.ConnOption) error {
	httpConn, err := http.NewClientConn(ctx, target, opts...)
	if err != nil {
		return err
	}
	cc.httpConn = httpConn
	return nil
}

// HttpInvoke http invoke
func (cc *ClientConn) HttpInvoke(ctx context.Context, method string, api string, req interface{}, reply interface{}, opts ...http.CallOption) error {
	if cc.httpConn == nil {
		return errors.New("httpClientConn empty, please init it")
	}
	return cc.httpConn.Invoke(ctx, method, api, req, reply, opts...)
}

// GrpcInvoke grpc invoke
func (cc *ClientConn) GrpcInvoke(ctx context.Context, method string, req interface{}, reply interface{}, opts ...grpc.CallOption) error {
	if cc.grpcConn == nil {
		return errors.New("grpcClientConn empty, please init it")
	}
	return cc.grpcConn.Invoke(ctx, method, req, reply, convertToRawGrpcCallOption(opts...)...)
}

func convertToRawGrpcCallOption(opts ...grpc.CallOption) []grpcRaw.CallOption {
	newOpts := make([]grpcRaw.CallOption, len(opts))
	for i, opt := range opts {
		newOpts[i] = opt
	}
	return newOpts
}

// NewStream grpc NewStream invoke
func (cc *ClientConn) NewStream(ctx context.Context, desc *grpcRaw.StreamDesc, method string, opts ...grpc.CallOption) (grpcRaw.ClientStream, error) {
	if cc.grpcConn == nil {
		return nil, errors.New("grpcClientConn empty, please init it")
	}
	return cc.grpcConn.NewStream(ctx, desc, method, convertToRawGrpcCallOption(opts...)...)
}
