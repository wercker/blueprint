# blueprint

Blueprint takes a template and turns it into a application. And more!

Blueprint is now a slightly neater system for managing the boilerplate for our
services. It requires a couple things to be set up correctly (but it should
set that up itself if it created the project), but it should allow us to fairly
trivially keep the boilerplate of our applications up to date.

# Design

This is how a new project will be set up, we'll need to make existing projects
look enough like this for the rest of the magic to work.

file structure from blueprint root:

```
- templates           # where templates are held
  - service           # a template named service
- managed             # where services managed by blueprint live
  - inspector         # a service managed by blueprint
    - .managed.json   # the config used to generate this service
```

git branch structure for a managed project
```
blueprint
  \-- master
      \-- short-lived branches
```

TODO(termie): deleting old files?

# Updating

Our main flow will be to:

  1. Update remote branches for managed service `$service`
  2. Switch to `blueprint` branch
  3. Create a new branch called `blueprint_update`
  4. Update / overwrite the contents of the directory with the new code based
     on existing `.managed.json`
  5. Automatically (?) commit it and show a diff against the `blueprint` branch
  6. Continue y/n
  7. Switch to the `master` branch
  8. Create a new branch called `master_update`
  9. `rebase -i blueprint_update`, try to do the rebase
  10. If it all works, hard merge over master and ff-merge `blueprint_update` to
      `blueprint`

# Usage

Start a new project:

`blueprint init service $name`

Update a project that is already checked out in ./managed/$name:

`blueprint apply service $name`




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
  6666 -> {{.Port}}
  6667 -> {{.Gateway}}
  1996 -> {{.Year}}
  Tivo for VRML -> {{.Description}}
```
