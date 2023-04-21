package grpc

import (
	"context"
	"github.com/classtorch/prpc/balancer/roundrobin"
	"github.com/classtorch/prpc/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

func Test_NewClientConn(t *testing.T) {
	ctx := context.Background()
	// direct
	client, err := NewClientConn(context.Background(), "127.0.0.1:33000", WithOptions(grpc.WithTransportCredentials(insecure.NewCredentials())))
	if err != nil {
		t.Error(err)
	}
	err = client.Invoke(ctx, "getAccountList", nil, nil)
	if err != nil {
		t.Error(err)
	}
	// resolver
	client, err = NewClientConn(context.Background(), "etcd://127.0.0.1:8000/account", WithResolver(mockResolverBuilder{}), WithOptions(grpc.WithTransportCredentials(insecure.NewCredentials())))
	if err != resolver.ResolverNotExistErr {
		t.Errorf("expect err:ResolverNotExistErr but get:%v", err)
	}
	client, err = NewClientConn(context.Background(), "consul://127.0.0.1:8000/account", WithResolver(mockResolverBuilder{}), WithOptions(grpc.WithTransportCredentials(insecure.NewCredentials())))
	if err != nil {
		t.Error(err)
	}
	err = client.Invoke(ctx, "getAccountList", nil, nil)
	if err != nil {
		t.Error(err)
	}
	// resolver,balancer
	client, err = NewClientConn(context.Background(), "consul://127.0.0.1:8000/account", WithResolver(mockResolverBuilder{}), WithBalancerName(roundrobin.Name), WithOptions(grpc.WithTransportCredentials(insecure.NewCredentials())))
	if err != nil {
		t.Error(err)
	}
	err = client.Invoke(ctx, "getAccountList", nil, nil)
	if err != nil {
		t.Error(err)
	}
}
