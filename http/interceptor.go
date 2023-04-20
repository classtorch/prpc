package http

import (
	"context"
	"net/http"
)

type Invoker func(ctx context.Context, req interface{}, reply interface{}, httpRequest *http.Request, httpResponse *http.Response, cc *ClientConn, option ...CallOption) error

type Interceptor func(ctx context.Context, req interface{}, reply interface{}, httpRequest *http.Request, httpResponse *http.Response, cc *ClientConn, invoker Invoker, option ...CallOption) error

func ChainUnaryInterceptors(cc *ClientConn) {
	interceptors := cc.connOption.chanInterceptor
	if cc.connOption.unaryInterceptor != nil {
		interceptors = append([]Interceptor{cc.connOption.unaryInterceptor}, interceptors...)
	}
	if len(interceptors) > 0 {
		cc.connOption.unaryInterceptor = func(ctx context.Context, req interface{}, reply interface{}, httpRequest *http.Request, httpResponse *http.Response, cc *ClientConn, invoker Invoker, option ...CallOption) error {
			return interceptors[0](ctx, req, reply, httpRequest, httpResponse, cc, GetChainUnaryInvoker(interceptors, 0, invoker), option...)
		}
	}
}

func GetChainUnaryInvoker(interceptors []Interceptor, cur int, finalInvoker Invoker) Invoker {
	if cur == len(interceptors)-1 {
		return finalInvoker
	}
	return func(ctx context.Context, req interface{}, reply interface{}, httpRequest *http.Request, httpResponse *http.Response, cc *ClientConn, option ...CallOption) error {
		return interceptors[cur+1](ctx, req, reply, httpRequest, httpResponse, cc, GetChainUnaryInvoker(interceptors, cur+1, finalInvoker), option...)
	}
}
