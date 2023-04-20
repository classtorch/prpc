##### Translate to: [简体中文](README_zh.md)

# proto-gen-go-prpc

This is a http and grpc code generation tool defined in protobuf, containing:

* Generate http client code
* Generate grpc client code and server code, fully compatible with grpc official grpc-go code generation tool [protoc-gen-go-grpc](https://github.com/grpc/grpc-go/tree/master/cmd/protoc-gen-go-grpc)
```
Install：
go install https://github.com/classtorch/prpc/cmd/protoc-gen-go-prpc

Use：
protoc --go-prpc_out=:. --go_out=.  -I . api/user.proto
```

## Related

* [Examples](https://github.com/classtorch/prpc-examples)
