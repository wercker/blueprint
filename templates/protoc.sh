#!/bin/sh

echo "protobuf: $PROTOBUF_VERSION"
echo "protoc-gen-grpc-gateway: $PROTOC_GRPC_VERSION"
echo "protoc-gen-swagger: $PROTOC_GRPC_VERSION"
echo "protoc-gen-go: $PROTOC_GO_VERSION"
echo "protoc-gen-flow: $PROTOC_FLOW_VERSION"

protoc \
  -I. \
  -I/usr/local/include \
  -I/go/src \
  -I./vendor \
  -I./vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
  \$@
