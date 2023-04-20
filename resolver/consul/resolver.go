package consul

import (
	"context"
	"fmt"
	"github.com/classtorch/prpc/resolver"
	"github.com/hashicorp/consul/api"
	"github.com/jpillora/backoff"
	"log"
	"time"
)

//go:generate mockgen -package mocks -destination internal/mocks/resolverClientConn.go  google.golang.org/grpc/resolver ClientConn
//go:generate mockgen -package mocks -destination internal/mocks/servicer.go -source consul.go servicer
type servicer interface {
	Service(string, string, bool, *api.QueryOptions) ([]*api.ServiceEntry, *api.QueryMeta, error)
}

func watchConsulService(ctx context.Context, s servicer, tgt target, out chan<- []string) {
	res := make(chan []string)
	quit := make(chan struct{})
	bck := &backoff.Backoff{
		Factor: 2,
		Jitter: true,
		Min:    10 * time.Millisecond,
		Max:    tgt.MaxBackoff,
	}
	go func() {
		var lastIndex uint64
		for {
			entries, meta, err := s.Service(
				tgt.Service,
				tgt.Tag,
				tgt.Healthy,
				&api.QueryOptions{
					WaitIndex:         lastIndex,
					Near:              tgt.Near,
					WaitTime:          tgt.Wait,
					Datacenter:        tgt.Dc,
					AllowStale:        tgt.AllowStale,
					RequireConsistent: tgt.RequireConsistent,
				},
			)
			if err != nil {
				log.Printf("[Consul resolver] Couldn't fetch endpoints. target={%s}; error={%v}", tgt.String(), err)
				time.Sleep(bck.Duration())
				continue
			}
			bck.Reset()
			lastIndex = meta.LastIndex
			log.Printf("[Consul resolver] %d endpoints fetched in(+wait) %s for target={%s}",
				len(entries),
				meta.RequestTime,
				tgt.String(),
			)

			addrs := make([]string, 0, len(entries))
			for _, s := range entries {
				address := s.Service.Address
				if s.Service.Address == "" {
					address = s.Node.Address
				}
				address = fmt.Sprintf("%s:%d", address, s.Service.Port)
				addrs = append(addrs, address)
			}

			if tgt.Limit != 0 && len(addrs) > tgt.Limit {
				addrs = addrs[:tgt.Limit]
			}
			select {
			case res <- addrs:
				continue
			case <-quit:
				return
			}
		}
	}()

	for {
		select {
		case ee := <-res:
			out <- ee
		case <-ctx.Done():
			// Close quit so the goroutine returns and doesn't leak.
			// Do NOT close res because that can lead to panics in the goroutine.
			// res will be garbage collected at some point.
			close(quit)
			return
		}
	}
}

func populateEndpoints(ctx context.Context, cc resolver.ClientConn, input <-chan []string) {
	for {
		select {
		case data := <-input:
			addresses := make([]resolver.Address, len(data))
			for i, addr := range data {
				addresses[i] = resolver.Address{Addr: addr}
			}
			cc.UpdateState(resolver.State{Addresses: addresses})
		case <-ctx.Done():
			log.Printf("[Consul resolver] Watch has been finished")
			return
		}
	}
}
