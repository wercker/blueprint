#!/bin/bash
set -e

(

# Add any required imports here, seperated by commas
CUSTOM_IMPORTS=""

LOCAL=$(dirname $PWD)

if [ "$FORCE_LOCAL" != "true" ] && [ -e /var/run/docker.sock ]; then
  ROOT=${LOCAL//$GOPATH/\/go}
  GEN="docker run --rm \
    -u $(id -u $USER):$(id -g $USER) \
    -w $ROOT \
    -v $LOCAL:$ROOT \
    quay.io/wercker/igenerator:stable"
  TEMPLATE_DIR="/go/src/github.com/wercker/blueprint/cmd/igenerator"
else
  GENERATOR_PATH=${GENERATOR_PATH:?"GENERATOR_PATH is required for local runs"}
  ROOT=$LOCAL
  GEN="go run $GENERATOR_PATH/main.go"
  TEMPLATE_DIR="$GENERATOR_PATH"
fi

echo "Generating metrics store"
$GEN -target=Store \
     -input="$ROOT/state/store.go" \
     -ignore="Initialize" \
     -template="$TEMPLATE_DIR/metrics_store.go.tpl" \
     -output="$ROOT/state/metrics_store.go" \
     -imports="$CUSTOM_IMPORTS"

echo "Generating trace store"
$GEN -target=Store \
     -input="$ROOT/state/store.go" \
     -ignore="Initialize" \
     -template="$TEMPLATE_DIR/trace_store.go.tpl" \
     -output="$ROOT/state/trace_store.go" \
     -imports="$CUSTOM_IMPORTS"

)
