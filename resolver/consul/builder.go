package consul

import (
	"context"
	"strings"

	"github.com/classtorch/prpc/resolver"
	"github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
)

// schemeName for the urls
// All target URLs like 'consul://.../...' will be resolved by this resolver
const schemeName = "consul"

// new a resolver builder
func NewResolverBuilder() resolver.Builder {
	return &builder{}
}

// builder implements resolver.Builder and use for constructing all consul resolvers
type builder struct{}

func (b *builder) Build(target resolver.Target, cc resolver.ClientConn) (resolver.Resolver, error) {
	dsn := strings.Join([]string{schemeName + ":/", target.Agent, target.Endpoint}, "/")
	tgt, err := parseURL(dsn)
	if err != nil {
		return nil, errors.Wrap(err, "Wrong consul URL")
	}
	cli, err := api.NewClient(tgt.consulConfig())
	if err != nil {
		return nil, errors.Wrap(err, "Couldn't connect to the Consul API")
	}

	ctx, cancel := context.WithCancel(context.Background())
	pipe := make(chan []string)
	go watchConsulService(ctx, cli.Health(), tgt, pipe)
	go populateEndpoints(ctx, cc, pipe)

	return &Resolver{cancelFunc: cancel}, nil
}

// Scheme returns the scheme supported by this resolver.
// Scheme is defined at https://github.com/grpc/grpc/blob/master/doc/naming.md.
func (b *builder) Scheme() string {
	return schemeName
}

// resolvr implements resolver.Resolver from the gRPC package.
// It watches for endpoints changes and pushes them to the underlying gRPC connection.
type Resolver struct {
	cancelFunc context.CancelFunc
}

// ResolveNow will be skipped due unnecessary in this case
func (r *Resolver) ResolveNow() {}

// Close closes the resolver.
func (r *Resolver) Close() {
	r.cancelFunc()
}
