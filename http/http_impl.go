package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/go-querystring/query"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

const (
	ContentType     = "Content-Type"
	ContentTypeForm = "application/x-www-form-urlencoded"
	ContentTypeJson = "application/json;charset=utf-8"
)

// CallInterface http client impl interface
type CallInterface interface {
	Get(ctx context.Context, addr string, api string, req interface{}, reply interface{}, opts ...CallOption) (*http.Request, *http.Response, error)
	Post(ctx context.Context, addr string, api string, req interface{}, reply interface{}, opts ...CallOption) (*http.Request, *http.Response, error)
	Put(ctx context.Context, addr string, api string, req interface{}, reply interface{}, opts ...CallOption) (*http.Request, *http.Response, error)
	Delete(ctx context.Context, addr string, api string, req interface{}, reply interface{}, opts ...CallOption) (*http.Request, *http.Response, error)
	Default(ctx context.Context, addr string, api string, req interface{}, reply interface{}, opts ...CallOption) (*http.Request, *http.Response, error)
}

// defaultHttpClient default http client impl
type defaultHttpClient struct {
}

func NewDefaultPRpcHttpClient() CallInterface {
	return &defaultHttpClient{}
}

func (cc *defaultHttpClient) Get(ctx context.Context, addr string, api string, req interface{}, reply interface{}, opts ...CallOption) (*http.Request, *http.Response, error) {
	values, err := query.Values(req)
	if err != nil {
		return nil, nil, err
	}
	queryParams := values.Encode()
	url := addr + api + "?" + queryParams
	request, err := getRequest(ctx, url, http.MethodGet, nil, opts...)
	if err != nil {
		return nil, nil, err
	}
	response, err := do(request, CallOptions(opts).GetTimeOut(), reply)
	return request, response, err
}

func (cc *defaultHttpClient) Post(ctx context.Context, addr string, api string, req interface{}, reply interface{}, opts ...CallOption) (*http.Request, *http.Response, error) {
	return cc.doBodyRequest(ctx, addr, api, http.MethodPost, req, reply, opts...)
}

func (cc *defaultHttpClient) Delete(ctx context.Context, addr string, api string, req interface{}, reply interface{}, opts ...CallOption) (*http.Request, *http.Response, error) {
	return cc.doBodyRequest(ctx, addr, api, http.MethodDelete, req, reply, opts...)
}

func (cc *defaultHttpClient) Put(ctx context.Context, addr string, api string, req interface{}, reply interface{}, opts ...CallOption) (*http.Request, *http.Response, error) {
	return cc.doBodyRequest(ctx, addr, api, http.MethodPut, req, reply, opts...)
}

func (cc *defaultHttpClient) Default(ctx context.Context, addr string, api string, req interface{}, reply interface{}, opts ...CallOption) (*http.Request, *http.Response, error) {
	return cc.doBodyRequest(ctx, addr, api, http.MethodDelete, req, reply, opts...)
}

func (cc *defaultHttpClient) doBodyRequest(ctx context.Context, addr string, api string, method string, req interface{}, reply interface{}, opts ...CallOption) (*http.Request, *http.Response, error) {
	var reader io.Reader
	if checkPostFrom(opts...) {
		params, err := getPostFormParams(req)
		if err != nil {
			return nil, nil, err
		}
		reader = strings.NewReader(params.Encode())
	} else {
		bys, err := json.Marshal(req)
		if err != nil {
			return nil, nil, err
		}
		reader = bytes.NewReader(bys)
	}
	url := addr + api
	request, err := getRequest(ctx, url, method, reader, opts...)
	if err != nil {
		return nil, nil, err
	}
	response, err := do(request, CallOptions(opts).GetTimeOut(), reply)
	return request, response, err
}

// getRequest return http Request
func getRequest(ctx context.Context, url, method string, reader io.Reader, opts ...CallOption) (*http.Request, error) {
	request, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return nil, err
	}
	header := CallOptions(opts).GetHeader()
	if _, ok := header[ContentType]; !ok {
		header[ContentType] = ContentTypeJson
	}
	for k, v := range header {
		request.Header.Set(k, v)
	}
	return request, nil
}

// getPostFormParams convert req to form format params
func getPostFormParams(req interface{}) (url.Values, error) {
	params := make(url.Values)
	value := reflect.ValueOf(req)
	if value.Kind() != reflect.Struct {
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return nil, errors.New("req not struct")
	}
	n := value.NumField()
	valType := value.Type()
	for i := 0; i < n; i++ {
		field := valType.Field(i)
		val := value.Field(i)
		tagValue := field.Tag.Get("json")
		if len(tagValue) > 0 {
			tagValues := strings.Split(tagValue, ",")
			if len(tagValues) >= 1 {
				valStr := fmt.Sprintf("%v", val.Interface())
				params[tagValues[0]] = []string{valStr}
			}
		}
	}
	return params, nil
}

// checkPostFrom determine whether the request is a post form
func checkPostFrom(opts ...CallOption) bool {
	isForm := false
	for k, v := range CallOptions(opts).GetHeader() {
		if k == ContentType && v == ContentTypeForm {
			isForm = true
			break
		}
	}
	return isForm
}

// do execute request
func do(request *http.Request, timeOut time.Duration, reply interface{}) (*http.Response, error) {
	client := http.Client{Timeout: timeOut}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	if resp != nil {
		defer resp.Body.Close()
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if respBytes == nil || len(respBytes) == 0 {
		return nil, errors.New("response empty")
	}
	err = json.Unmarshal(respBytes, reply)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
