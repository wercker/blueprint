#!/bin/bash

set -e

ROOT=$(dirname $BASH_SOURCE)

OLD=$1
NEW=$2

if [[ -z "$OLD" || -z "$NEW" ]]; then
  echo "Need at least two args: new_pb.sh <old> <new> [<dir> <dir>]"
  exit 1
fi

# From here on we want to take a list of args
shift
shift


function replace_dir {
  searchdir=$1
  for x in $searchdir/*; do
    if [ -f "$x" ] && grep -q "core" $x; then
      echo "Replacing core -> ${NEW}pb in $x"
      $fake sed -i "s/core/${NEW}pb/g" $x
    fi
  done
}

#fake="echo"

echo "Transforming $OLD to $NEW..."
set -v
$fake git mv core/$OLD.proto ./$NEW.proto
$fake git rm -f core/$OLD.*
$fake git mv core ${NEW}pb
$fake sed -i s/core/${NEW}pb/g ${NEW}pb/proto.go
$fake sed -i s/$OLD\\\\/${NEW}\\\\/g wercker.yml
$fake sed "s/blue_print/$NEW/g" $ROOT/templates/service/blue_printpb/generate-protobuf.sh > ${NEW}pb/generate-protobuf.sh


while (( "$#" )); do
  replace_dir $1
  shift
done

set +v
