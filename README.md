# blueprint

Blueprint takes a template and turns it into a application.


# Building Protobuf 3 Locally

If you're on mac you'll need to `brew install automake` if you haven't already

```
  mkdir tmp
  cd tmp
  git clone https://github.com/google/protobuf
  cd protobuf
  ./autogen.sh
  ./configure --prefix /usr/local/Cellar/protobuf/3.0.0-dev
  make
  make install
  brew switch protobuf 3.0.0-dev
```


You're going to need the plugins installed globally (the binaries in your path)

```
  go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
  go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
  go get -u github.com/golang/protobuf/protoc-gen-go
```
# Generating the protobufs

We use go generate

```
govendor generate +local
```



# How to Write Templates

We're using some special sentinel values that we will find replace in the doc
before doing the templating:

```
  blueprint/templates/service -> {{package .Name}}
  Blueprint -> {{title .Name}}
  blueprint -> {{lower .Name}}
  666 -> {{.Port}}
  667 -> {{.Gateway}}
  1996 -> {{.Year}}
  Tivo for VRML -> {{.Description}}
```
