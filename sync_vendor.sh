#!/bin/bash

# This script syncs the versions of libraries in the current vendor to the
# "official" versions in the blueprint template source.

# Example usage:
#  cd $GOPATH/src/github.com/wercker/trigger
#  ../blueprint/sync_vendor.sh

ROOT=$(dirname $BASH_SOURCE)
jq '.package[] | "\(.path)@\(.revision)"' < $ROOT/templates/service/vendor/vendor.json | xargs -n 1 govendor fetch
