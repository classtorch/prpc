##### Translate to: [简体中文](README_zh.md)

# pRPC

pRPC is a golang http and grpc protocol client library driven by protobuf, and supports service discovery and load balancing

## Architecture
![image](https://github.com/classtorch/prpc/blob/master/prpc.png)

## Goals

Unifiy http and grpc calls,reduce the code workload of the interface in the process of switching between http and grpc protocols, and try to make users insensitive to the protocol at the code layer

## Principle

* Tool priority: All codes that can be generated by tools are generated by tools, reducing the error rate of handwritten codes and improving development efficiency;
* Pluggable: increase scalability through good interface design;


## Features
* Generate http and grpc calling codes defined in protobuf through the [protoc-gen-go-prpc](https://github.com/classtorch/prpc/blob/master/cmd/protoc-gen-go-prpc/README.md) tool
* The http client implementation supports extensions;
* HTTP and grpc client request services support service resolver and load balancing functions, and support expansion;
* Consul service discovery and the load balancing algorithm of  and polling is implemented, and others can be expanded according to needs;
* Support interceptors, which can easily implement log, trace and other functions;

## Getting Started
### Required
- [go](https://golang.org/dl/)
- [protoc](https://github.com/protocolbuffers/protobuf)
- [protoc-gen-go](https://github.com/protocolbuffers/protobuf-go)

### Installing
##### code generation tool protoc-gen-go-prpc installation and use：
```
Install：
go install github.com/classtorch/prpc/cmd/protoc-gen-go-prpc

Use：
protoc --go-prpc_out=:. --go_out=.  -I . api/user.proto
```

## Related

* [Examples](https://github.com/classtorch/prpc-examples)
