/*
 *
 * Copyright 2017 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package wrapper

import (
	"github.com/classtorch/prpc/resolver"
	"google.golang.org/grpc/balancer"
	"sync"
)

// CCResolverWrapper is a wrapper on top of cc for resolvers.
// It implements resolver.ClientConn interface.
type CCResolverWrapper struct {
	cc         resolver.ClientConnResolverInterface
	resolverMu sync.Mutex
	resolver   resolver.Resolver
	curState   resolver.State

	incomingMu sync.Mutex // Synchronizes all the incoming calls.
}

// NewCCResolverWrapper uses the resolver.Builder to build a Resolver and
// returns a CCResolverWrapper object which wraps the newly built resolver.
func NewCCResolverWrapper(cc resolver.ClientConnResolverInterface, rb resolver.Builder) (*CCResolverWrapper, error) {
	ccr := &CCResolverWrapper{
		cc: cc,
	}

	var err error
	// We need to hold the lock here while we assign to the ccr.resolver field
	// to guard against a data race caused by the following code path,
	// rb.Build-->ccr.ReportError-->ccr.poll-->ccr.resolveNow, would end up
	// accessing ccr.resolver which is being assigned here.
	ccr.resolverMu.Lock()
	defer ccr.resolverMu.Unlock()
	ccr.resolver, err = rb.Build(cc.GetParsedTarget(), ccr)
	if err != nil {
		return nil, err
	}
	return ccr, nil
}

func (ccr *CCResolverWrapper) ResolveNow() {
	ccr.resolverMu.Lock()
	ccr.resolver.ResolveNow()
	ccr.resolverMu.Unlock()
}

func (ccr *CCResolverWrapper) Close() {
	ccr.resolverMu.Lock()
	ccr.resolver.Close()
	ccr.resolverMu.Unlock()
}

// UpdateState update balancer State
func (ccr *CCResolverWrapper) UpdateState(s resolver.State) error {
	ccr.incomingMu.Lock()
	defer ccr.incomingMu.Unlock()
	ccr.curState = s
	if err := ccr.cc.UpdateResolverState(ccr.curState, nil); err == balancer.ErrBadResolverState {
		return balancer.ErrBadResolverState
	}
	return nil
}

// ReportError report a error
func (ccr *CCResolverWrapper) ReportError(err error) {
	ccr.incomingMu.Lock()
	defer ccr.incomingMu.Unlock()
	ccr.cc.UpdateResolverState(resolver.State{}, err)
}
