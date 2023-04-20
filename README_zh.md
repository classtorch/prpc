Translations: [English](README.md) | [简体中文](README_zh.md)

# pRPC

pRPC是一个由protoBuffer文件驱动的http和grpc协议客户端库，并支持服务发现和负载均衡 

## 架构图
![image](https://github.com/classtorch/prpc/blob/master/prpc.png)

## 目标

减少接口在http和grpc协议切换过程中代码工作量，尽量让使用者在代码层对协议无感

## 原则

* 工具优先：能通过工具生成的代码都用工具生成，减少手写代码错误率并提高开发效率；
* 可插拔：通过良好的接口设计来增加扩展性；


## 功能
* 通过 [protoc-gen-go-prpc](https://github.com/classtorch/prpc/blob/master/cmd/protoc-gen-go-prpc/README.md) 工具生成在protobuf中定义http和grpc调用代码
* ProtoBuffer单个Service支持同时定义http和grpc方法；
* http客户端实现支持扩展；
* http和grpc客户端请求服务支持服务解析和负载均衡功能，支持扩展；
* 实现consul服务发现和轮询的负载均衡算法，其他可根据需要自行扩展；
* 支持拦截器，可轻松实现日志、trace等功能；

## 快速开始
### 需要
- [go](https://golang.org/dl/)
- [protoc](https://github.com/protocolbuffers/protobuf)
- [protoc-gen-go](https://github.com/protocolbuffers/protobuf-go)

### 安装
##### 代码生成工具 protoc-gen-go-prpc 安装和使用：
```
安装：
go install github.com/classtorch/prpc/cmd/protoc-gen-go-prpc

使用：
protoc --go-prpc_out=:. --go_out=.  -I . api/user.proto
```

## 相关链接

* [Examples](https://github.com/classtorch/prpc-examples)


