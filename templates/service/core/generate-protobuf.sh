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
  --grpc-gateway_out=logtostderr=true,request_context=true:. \
  blueprint.proto

echo "Generating swagger"
protoc -I/usr/local/include \
  -I. \
  -I$GOPATH/src \
  -I../vendor \
  -I../vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
  --swagger_out=logtostderr=true:. \
  blueprint.proto
