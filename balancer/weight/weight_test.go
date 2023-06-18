package weight

import (
	"github.com/classtorch/prpc/balancer"
	"github.com/classtorch/prpc/resolver"
	"testing"
	"time"
)

func Test_Picker(t *testing.T) {
	rb := &rbPickerBuilder{}
	readyAddresses := []resolver.Address{
		{Addr: "127.0.0.1:8001", Attributes: map[interface{}]interface{}{"weight": 1}},
		{Addr: "127.0.0.1:8002", Attributes: map[interface{}]interface{}{"weight": 2}},
		{Addr: "127.0.0.1:8003", Attributes: map[interface{}]interface{}{"weight": 5}},
	}
	picker := rb.Build(balancer.PickerBuildInfo{ReadyAddresses: readyAddresses})
	for {
		result, err := picker.Pick()
		if err != nil {
			t.Fatal(err)
		}
		t.Log(result)
		time.Sleep(time.Second)
	}
}

func Test_Weight(t *testing.T) {
	rb := &RoundRobin{}
	rb.Add("127.0.0.1:8001", 1)
	rb.Add("127.0.0.1:8002", 2)
	rb.Add("127.0.0.1:8003", 5)
	for {
		t.Log(rb.Next())
		time.Sleep(time.Second)
	}
}
