Translations: [English](README.md) | [简体中文](README_zh.md)

# proto-gen-go-prpc

这是一个在protobuf中定义的http和grpc代码生成工具，包含：

* 生成http客户端代码
* 生成grpc客户端代码和服务端代码，完全兼容grpc官方的grpc-go的代码生成工具 [protoc-gen-go-grpc](https://github.com/grpc/grpc-go/tree/master/cmd/protoc-gen-go-grpc)

```
安装：
go install https://github.com/classtorch/prpc/cmd/protoc-gen-go-prpc

使用：
protoc --go-prpc_out=:. --go_out=.  -I . api/user.proto
```

## 相关链接

* [Examples](https://github.com/classtorch/prpc-examples)


