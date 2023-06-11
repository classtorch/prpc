package http

import (
	"strings"
	"testing"
)

func Test_GetVariableUrlParams(t *testing.T) {
	testCases := []struct {
		input  string
		expect []string
	}{
		{
			input:  "users/{uid}/order/{order_id}",
			expect: []string{"uid", "order_id"},
		},
		{
			input:  "users/{uid}}",
			expect: []string{"uid"},
		},
		{
			input:  "users",
			expect: []string{},
		},
	}

	for _, tCase := range testCases {
		results := getVariableUrlParams(tCase.input)
		if strings.Join(results, "") != strings.Join(tCase.expect, "") {
			t.Fatalf("expect:%v,but get:%v", tCase.expect, results)
		}
	}
	t.Log("success")
}

func Test_ConvertApi(t *testing.T) {
	testCases := []struct {
		api          string
		urlParam     map[string]string
		expectErr    bool
		expectResult string
	}{
		{
			api: "users/{uid}/orders/{order_id}",
			urlParam: map[string]string{
				"order_id": "order001",
				"uid":      "user001",
			},
			expectErr:    false,
			expectResult: "users/user001/orders/order001",
		},
		{
			api: "users/{uid}/orders",
			urlParam: map[string]string{
				"uid": "user001",
			},
			expectErr:    false,
			expectResult: "users/user001/orders",
		},
		{
			api: "users/order",
			urlParam: map[string]string{
				"uid": "user001",
			},
			expectErr:    false,
			expectResult: "users/order",
		},
		{
			api: "users/{uid}/orders",
			urlParam: map[string]string{
				"order_id": "order001",
			},
			expectErr:    true,
			expectResult: "",
		},
		{
			api: "users/{uid}/orders",
			urlParam: map[string]string{
				"uid": "",
			},
			expectErr:    true,
			expectResult: "",
		},
	}
	for _, tCase := range testCases {
		result, err := convertApi(tCase.api, tCase.urlParam)
		if tCase.expectErr && err == nil {
			t.Fatalf("expect has err ,but get nil")
		}
		if tCase.expectResult != result {
			t.Fatalf("expect:%v,but get:%v", tCase.expectResult, result)
		}
	}
	t.Log("success")
}
