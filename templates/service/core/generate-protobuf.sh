#!/bin/bash

set -e

echo "Generating gRPC server"
protoc -I/usr/local/include \
  -I. \
  -I$GOPATH/src \
  -I../vendor \
  -I../vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
  --go_out=Mgoogle/api/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api,plugins=grpc:. \
  blueprint.proto

echo "Generating gateway"
protoc -I/usr/local/include \
  -I. \
  -I$GOPATH/src \
  -I../vendor \
  -I../vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
  --grpc-gateway_out=logtostderr=true:. \
  blueprint.proto

echo "Generating swagger"
protoc -I/usr/local/include \
  -I. \
  -I$GOPATH/src \
  -I../vendor \
  -I../vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
  --swagger_out=logtostderr=true:. \
  blueprint.proto

# This hack is to ensure that we're using the context provided by the request
echo "Applying context hack to gateway"
if [[ "$(uname)" == "Darwin" ]]; then
    sed -i '' 's/ctx, cancel := context\.WithCancel(ctx)/ctx, cancel := context.WithCancel(req.Context())/g' blueprint.pb.gw.go
else
    sed -i 's/ctx, cancel := context\.WithCancel(ctx)/ctx, cancel := context.WithCancel(req.Context())/g' blueprint.pb.gw.go
