package http

import (
	"context"
	"testing"
)

type GetUserInfoReq struct {
	Uid  uint32 `json:"uid,omitempty"`
	Name string `json:"name,omitempty"`
}

type GetUserInfoReply struct {
}

func Test_Get(t *testing.T) {
	client := NewDefaultPRpcHttpClient()
	_, _, err := client.Get(context.Background(), "https://127.0.0.1:8000/", "users", &GetUserInfoReq{Uid: 123456, Name: "张三"}, &GetUserInfoReply{})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("success")
}
