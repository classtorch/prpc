package main

import (
	"flag"
	"fmt"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

var showVersion = flag.Bool("version", false, "print the version and exit")
var requireUnimplemented *bool
var httpGenerateGrpc = flag.Bool("http_generate_grpc", true, "set whether http needs to generate grpc methods")

const version = "v1.0.1"

func main() {
	flag.Parse()
	if *showVersion {
		fmt.Printf("protoc-gen-go-prpc %s\n", version)
		return
	}
	var flags flag.FlagSet
	requireUnimplemented = flags.Bool("require_unimplemented_servers", true, "set to false to match legacy behavior")

	protogen.Options{
		ParamFunc: flag.CommandLine.Set,
	}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			generateFileGrpc(gen, f)
		}
		return nil
	})
}
