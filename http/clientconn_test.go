package http

import (
	"context"
	"fmt"
	"github.com/classtorch/prpc/balancer/roundrobin"
	"github.com/classtorch/prpc/resolver"
	"net/http"
	"strings"
	"testing"
)

type mockResolverBuilder struct {
}

var services = []string{"127.0.0.1:33000", "127.0.0.2:33000", "127.0.0.3:33000"}

func (resolverBuilder mockResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn) (resolver.Resolver, error) {
	addresses := make([]resolver.Address, len(services))
	for i, addr := range services {
		addresses[i] = resolver.Address{Addr: addr}
	}
	state := resolver.State{Addresses: addresses}
	cc.UpdateState(state)
	return mockResolver{}, nil
}

func (resolverBuilder mockResolverBuilder) Scheme() string {
	return "consul"
}

type mockResolver struct {
}

func (mockResolver mockResolver) ResolveNow() {
}

func (mockResolver mockResolver) Close() {
}

type mockHttpImpl struct {
}

func (mockHttpImpl mockHttpImpl) Get(ctx context.Context, addr string, api string, req interface{}, reply interface{}, opts ...CallOption) (*http.Request, *http.Response, error) {
	request, err := http.NewRequest(http.MethodGet, addr+api, nil)
	if err != nil {
		return nil, nil, err
	}
	return request, &http.Response{}, nil
}

func (mockHttpImpl mockHttpImpl) Post(ctx context.Context, addr string, api string, req interface{}, reply interface{}, opts ...CallOption) (*http.Request, *http.Response, error) {
	return nil, nil, nil
}
func (mockHttpImpl mockHttpImpl) Delete(ctx context.Context, addr string, api string, req interface{}, reply interface{}, opts ...CallOption) (*http.Request, *http.Response, error) {
	return nil, nil, nil
}
func (mockHttpImpl mockHttpImpl) Put(ctx context.Context, addr string, api string, req interface{}, reply interface{}, opts ...CallOption) (*http.Request, *http.Response, error) {
	return nil, nil, nil
}
func (mockHttpImpl mockHttpImpl) Default(ctx context.Context, addr string, api string, req interface{}, reply interface{}, opts ...CallOption) (*http.Request, *http.Response, error) {
	return nil, nil, nil
}

func Test_NewClientConn(t *testing.T) {
	ctx := context.Background()
	// direct
	client, err := NewClientConn(context.Background(), "127.0.0.1:8000/account", WithCallClient(mockHttpImpl{}))
	if err != nil {
		t.Error(err)
	}
	err = client.Invoke(ctx, "GET", "/getUserList", nil, nil, WithUrlParam("/123"))
	if err != nil {
		t.Error(err)
	}
	// resolver
	client, err = NewClientConn(context.Background(), "etcd://127.0.0.1:8000/account", WithResolver(mockResolverBuilder{}), WithCallClient(mockHttpImpl{}))
	if err != resolver.ResolverNotExistErr {
		t.Errorf("expect err:ResolverNotExistErr but get:%v", err)
	}
	client, err = NewClientConn(context.Background(), "consul://127.0.0.1:8000/account", WithResolver(mockResolverBuilder{}), WithCallClient(mockHttpImpl{}))
	if err != nil {
		t.Error(err)
	}
	err = client.Invoke(ctx, "GET", "/getUserList", nil, nil)
	if err != nil {
		t.Error(err)
	}
	// resolver,balancer
	client, err = NewClientConn(context.Background(), "consul://127.0.0.1:8000/account", WithResolver(mockResolverBuilder{}), WithCallClient(mockHttpImpl{}), WithBalancerName(roundrobin.Name))
	if err != nil {
		t.Error(err)
	}
	err = client.Invoke(ctx, "GET", "/getUserList", nil, nil)
	if err != nil {
		t.Error(err)
	}
	// resolver,balancer,middleware
	index := 0
	client, err = NewClientConn(context.Background(), "consul://127.0.0.1:8000/account", WithResolver(mockResolverBuilder{}), WithCallClient(mockHttpImpl{}), WithBalancerName(roundrobin.Name), WithInterceptor(
		func(ctx context.Context, req interface{}, reply interface{}, httpRequest *http.Request, httpResponse *http.Response, cc *ClientConn, invoker Invoker, option ...CallOption) error {
			fmt.Println(httpRequest.Host)
			addr := strings.Split(httpRequest.Host, "//")[1]
			if addr != services[index] {
				t.Errorf("expect addr:%s but get:%s", services[index], httpRequest.Host)
			}
			index++
			return invoke(ctx, req, reply, httpRequest, httpResponse, cc, option...)
		}))
	if err != nil {
		t.Error(err)
	}
	for i := 0; i <= 2; i++ {
		err = client.Invoke(ctx, "GET", "/getUserList", nil, nil)
		if err != nil {
			t.Error(err)
		}
	}
}
