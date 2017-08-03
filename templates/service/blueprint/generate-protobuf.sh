#!/bin/bash
set -e

LOCAL=$(dirname $PWD)
ROOT=${LOCAL//$GOPATH/\/go}

protoc="docker run --rm \
  -u $(id -u $USER):$(id -g $USER) \
  -w $ROOT \
  -v $LOCAL:$ROOT \
  quay.io/wercker/protoc"

echo "Generating gRPC server, gateway, swagger"
$protoc --go_out=plugins=grpc:$ROOT/blueprint \
        --grpc-gateway_out=logtostderr=true,request_context=true:$ROOT/blueprint \
        --swagger_out=logtostderr=true:$ROOT/blueprint \
        blueprint.proto

echo "Generating flow types"
$protoc --flow_out=$ROOT/blueprint \
        blueprint.proto
