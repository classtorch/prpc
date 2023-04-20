package resolver

import (
	"errors"
	toolString "github.com/classtorch/prpc/pkg/strings"
	"google.golang.org/grpc/attributes"
	"strings"
)

var (
	m           = make(map[string]Builder)
	passThrough = "pass_through"
)

var (
	ResolverNotExistErr = errors.New("scheme resolver not exist")
)

// Register register a resolver Builder
func Register(b Builder) {
	m[b.Scheme()] = b
}

// Get get a resolver Builder by scheme
func Get(scheme string) Builder {
	if b, ok := m[scheme]; ok {
		return b
	}
	return nil
}

// GetPassThroughScheme get passThrough flag
func GetPassThroughScheme() string {
	return passThrough
}

// Address represents a server the client connects to.
//
// # Experimental
//
// Notice: This type is EXPERIMENTAL and may be changed or removed in a
// later release.
type Address struct {
	// Addr is the server address on which a connection will be established.
	Addr string

	// ServerName is the name of this address.
	// If non-empty, the ServerName is used as the transport certification authority for
	// the address, instead of the hostname from the Dial target string. In most cases,
	// this should not be set.
	//
	// If Type is GRPCLB, ServerName should be the name of the remote load
	// balancer, not the name of the backend.
	//
	// WARNING: ServerName must only be populated with trusted values. It
	// is insecure to populate it with data from untrusted inputs since untrusted
	// values could be used to bypass the authority checks performed by TLS.
	ServerName string

	// Attributes contains arbitrary data about this address intended for
	// consumption by the SubConn.
	Attributes map[interface{}]interface{}
}

// State contains the current Resolver state relevant to the ClientConn.
type State struct {
	// Addresses is the latest set of resolved addresses for the target.
	Addresses []Address
	// consumption by the load balancing policy.
	Attributes *attributes.Attributes
}

// Builder resolver Build interface
type Builder interface {
	Build(target Target, cc ClientConn) (Resolver, error)
	Scheme() string
}

// Resolver resolver interface
type Resolver interface {
	ResolveNow()
	Close()
}

// Target target such as "consul://127.0.0.1:8500/uclass-account"
// consul->schema
// 127.0.0.1:8500->agent
// uclass-account->endpoint
type Target struct {
	Scheme   string
	Agent    string
	Endpoint string
}

// ClientConn
type ClientConn interface {
	UpdateState(state State) error
	// ReportError notifies the ClientConn that the Resolver encountered an
	// error.  The ClientConn will notify the load balancer and begin calling
	// ResolveNow on the Resolver with exponential backoff.
	ReportError(error)
}

// ParseTarget convert string target to Target struct
func ParseTarget(target string) Target {
	var scheme = passThrough
	endpoint := ""
	agent := ""
	schemeIndex := strings.Index(target, "://")
	ok := false
	if schemeIndex > 0 {
		scheme, endpoint, ok = toolString.Split2(target, "://")
		if !ok {
			return Target{Endpoint: target}
		}
		agent, endpoint, ok = toolString.Split2(endpoint, "/")
		if !ok {
			return Target{Scheme: scheme, Endpoint: target}
		}
	}
	return Target{Scheme: scheme, Agent: agent, Endpoint: endpoint}
}

// ClientConnResolverInterface
type ClientConnResolverInterface interface {
	UpdateResolverState(s State, err error) error
	GetParsedTarget() Target
}
