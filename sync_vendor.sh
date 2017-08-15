#!/bin/bash
ROOT=$(dirname $BASH_SOURCE)
jq '.package[] | "\(.path)@\(.revision)"' < $ROOT/templates/service/vendor/vendor.json | xargs -n 1 govendor fetch
