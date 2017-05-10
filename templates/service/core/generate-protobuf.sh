#!/bin/bash

set -e

echo "Generating gRPC server, gateway, swagger"
protoc -I/usr/local/include \
  -I. \
  -I$GOPATH/src \
  -I../vendor \
  -I../vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
  --go_out=plugins=grpc:. \
  --grpc-gateway_out=logtostderr=true,request_context=true:. \
  --swagger_out=logtostderr=true:. \
  blueprint.proto
