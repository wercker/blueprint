package core

// gRPC Server
//go:generate protoc -I/usr/local/include -I. -I$GOPATH/src -I../vendor -I../vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --go_out=Mgoogle/api/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api,plugins=grpc:. blueprint.proto
//go:generate echo "gRPC server generated"

// gRPC Gateway
//go:generate protoc -I/usr/local/include -I. -I$GOPATH/src -I../vendor -I../vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --grpc-gateway_out=logtostderr=true:. blueprint.proto
//go:generate echo "gRPC gateway generated"

// Swagger
//go:generate protoc -I/usr/local/include -I. -I$GOPATH/src -I../vendor -I../vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --swagger_out=logtostderr=true:. blueprint.proto
//go:generate echo "gRPC gateway swagger generated"
