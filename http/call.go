package http

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"
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

func WithUrlParam(value string) CallOption {
	return func(callOption *callOption) {
		callOption.UrlParam = value
	}
}

type callOption struct {
	Header   map[string]string
	TimeOut  time.Duration
	UrlParam string // url param,if raw url is /users,url param='/123',then latest url is /users/123. the url param can be multiple，such as /123/001
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

func (opts CallOptions) GetUrlParam() string {
	callOpt := &callOption{}
	for _, opt := range opts {
		opt(callOpt)
	}
	return callOpt.UrlParam
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

// 整合clientConn和CallOption中的TimeOut值,如果CallOption没有配置，则取clientConn中的
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
	if len(callOpt.UrlParam) > 0 {
		options = append(options, WithUrlParam(callOpt.UrlParam))
	}
	return options
}

func invoke(ctx context.Context, req interface{}, reply interface{}, httpRequest *http.Request, httpResponse *http.Response, cc *ClientConn, opts ...CallOption) error {
	call := cc.GetOption().httpCall
	addr := httpRequest.Host
	api := httpRequest.URL.Path
	method := strings.ToUpper(httpRequest.Method)
	opts = combineCallOptions(cc, opts...)

	urlParam := CallOptions(opts).GetUrlParam()
	if len(urlParam) > 0 {
		api = api + urlParam
	}

	var err error
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
