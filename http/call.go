package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	variableUrlRex = regexp.MustCompile(`{(.*?)}`)
)

type CallOption func(callOption *callOption)

func WithHeader(header map[string]string) CallOption {
	return func(callOption *callOption) {
		callOption.Header = header
	}
}

func WithCallTimeOut(value int) CallOption {
	return func(callOption *callOption) {
		callOption.TimeOut = time.Second * time.Duration(value)
	}
}

func WithUrlParams(value map[string]string) CallOption {
	return func(callOption *callOption) {
		callOption.UrlParams = value
	}
}

type callOption struct {
	Header    map[string]string
	TimeOut   time.Duration
	UrlParams map[string]string // url params,if raw url is /users/{uid},url params=map{"uid":123},then latest url is /users/123.
}

type CallOptions []CallOption

func (opts CallOptions) CombineHeader(header map[string]string) []CallOption {
	callOpt := &callOption{}
	for _, opt := range opts {
		opt(callOpt)
	}
	if callOpt.Header == nil {
		callOpt.Header = map[string]string{}
	}
	for k, v := range header {
		if _, ok := callOpt.Header[k]; !ok {
			callOpt.Header[k] = v
		}
	}
	var options []CallOption
	options = append(options, WithHeader(callOpt.Header))
	if callOpt.TimeOut != 0 {
		options = append(options, WithCallTimeOut(int(callOpt.TimeOut/time.Second)))
	}
	if len(callOpt.UrlParams) > 0 {
		options = append(options, WithUrlParams(callOpt.UrlParams))
	}
	return options
}

func (opts CallOptions) GetTimeOut() time.Duration {
	callOpt := &callOption{}
	for _, opt := range opts {
		opt(callOpt)
	}
	return callOpt.TimeOut
}

func (opts CallOptions) GetHeader() map[string]string {
	callOpt := &callOption{}
	for _, opt := range opts {
		opt(callOpt)
	}
	if callOpt.Header == nil {
		callOpt.Header = map[string]string{}
	}
	return callOpt.Header
}

func (opts CallOptions) GetUrlParam() map[string]string {
	callOpt := &callOption{}
	for _, opt := range opts {
		opt(callOpt)
	}
	return callOpt.UrlParams
}

func getVariableUrlParams(url string) []string {
	params := variableUrlRex.FindAllString(url, -1)
	results := make([]string, len(params))
	for idx, temp := range params {
		result := strings.ReplaceAll(temp, "{", "")
		result = strings.ReplaceAll(result, "}", "")
		results[idx] = result
	}
	return results
}

func (cc *ClientConn) Invoke(ctx context.Context, method string, api string, req interface{}, reply interface{}, opts ...CallOption) error {
	addr := ""
	var err error
	if cc.direct {
		addr = cc.target
	} else {
		addr, err = cc.GetPickerWrapper().Pick(ctx, false)
		if err != nil {
			return err
		}
	}
	if cc.connOption.secure {
		addr = "https://" + addr
	} else {
		addr = "http://" + addr
	}
	request := &http.Request{Method: method, Host: addr, URL: &url.URL{Path: api}}

	if cc.GetOption().unaryInterceptor != nil {
		return cc.GetOption().unaryInterceptor(ctx, req, reply, request, &http.Response{}, cc, invoke, opts...)
	}
	return invoke(ctx, req, reply, request, &http.Response{}, cc, opts...)
}

func invoke(ctx context.Context, req interface{}, reply interface{}, httpRequest *http.Request, httpResponse *http.Response, cc *ClientConn, opts ...CallOption) error {
	call := cc.GetOption().httpCall
	addr := httpRequest.Host
	api := httpRequest.URL.Path
	method := strings.ToUpper(httpRequest.Method)
	opts = combineCallOptions(cc, opts...)

	urlParam := CallOptions(opts).GetUrlParam()
	var err error
	api, err = convertApi(api, urlParam)
	if err != nil {
		return err
	}
	switch method {
	case http.MethodGet:
		httpRequest, httpResponse, err = call.Get(ctx, addr, api, req, reply, opts...)
	case http.MethodPost:
		httpRequest, httpResponse, err = call.Post(ctx, addr, api, req, reply, opts...)
	case http.MethodPut:
		httpRequest, httpResponse, err = call.Put(ctx, addr, api, req, reply, opts...)
	case http.MethodDelete:
		httpRequest, httpResponse, err = call.Delete(ctx, addr, api, req, reply, opts...)
	default:
		httpRequest, httpResponse, err = call.Default(ctx, addr, api, req, reply, opts...)
	}
	return err
}

// combineCallOptions integrate the TimeOut value in clientConn and CallOption, if CallOption is not configured, take the value in clientConn
func combineCallOptions(cc *ClientConn, option ...CallOption) []CallOption {
	callOpt := &callOption{}
	for _, opt := range option {
		opt(callOpt)
	}
	if callOpt.TimeOut == 0 {
		callOpt.TimeOut = cc.connOption.timeOut
	}
	var options []CallOption
	if callOpt.Header != nil {
		options = append(options, WithHeader(callOpt.Header))
	}
	if callOpt.TimeOut != 0 {
		options = append(options, WithCallTimeOut(int(callOpt.TimeOut/time.Second)))
	}
	if len(callOpt.UrlParams) > 0 {
		options = append(options, WithUrlParams(callOpt.UrlParams))
	}
	return options
}

// convertApi Convert the variable parameter variable in the api address to a value
func convertApi(oldApi string, urlParam map[string]string) (string, error) {
	variableParams := getVariableUrlParams(oldApi)
	variableParamMap := make(map[string]string)
	newApi := oldApi
	if len(variableParams) == 0 {
		return newApi, nil
	}
	for _, key := range variableParams {
		paramVal, ok := urlParam[key]
		if !ok {
			return "", errors.New(fmt.Sprintf("variable url param key:%s not exist", key))
		}
		if len(paramVal) == 0 {
			return "", errors.New(fmt.Sprintf("variable url param key:%s's value can not empty", key))
		}
		variableParamMap[key] = paramVal
	}
	for key, value := range variableParamMap {
		newApi = strings.ReplaceAll(newApi, fmt.Sprintf("{%s}", key), value)
	}
	return newApi, nil
}
