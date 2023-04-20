package balancer

import (
	"errors"
	"github.com/classtorch/prpc/resolver"
	"strings"
)

var (
	ErrNoAddressAvailable = errors.New("no address is available")
	BalancerNotExistErr   = errors.New("special balancer not exist")
)

var (
	m = make(map[string]Builder)
)

func Get(name string) Builder {
	if builder, ok := m[name]; ok {
		return builder
	}
	return nil
}

func Register(builder Builder) {
	m[strings.ToLower(builder.Name())] = builder
}

// Builder creates a balancer.
type Builder interface {
	Build() Balancer
	Name() string
}

type Balancer interface {
	UpdateState(state resolver.State) (Picker, error)
	Close()
}

// PickerBuilder creates balancer.Picker.
type PickerBuilder interface {
	Build(info PickerBuildInfo) Picker
}

type PickerBuildInfo struct {
	ReadyAddresses []resolver.Address
}

type Picker interface {
	Pick() (PickResult, error)
}

type PickResult struct {
	Address resolver.Address
	Done    func(DoneInfo)
}

type DoneInfo struct {
	Err error
}
