#!/bin/bash

# This is a tool to migrate protocol buffers for a
# blueprint-like project from being in the `core/blueprint.proto`
# to being in `blueprint.proto` and putting the output into
# `blueprintpb/blueprint.pb.go`.

# Example usage output:
#termie@golfer:github.com/wercker/trigger % ../blueprint/migrate_pb.sh trigger trigger state queue providers events consumer clustermanager [pbswitch] 11:02:22
#Transforming trigger to trigger...
#$fake git mv core/$OLD.proto ./$NEW.proto
#$fake git rm -f core/$OLD.*
#rm 'core/trigger.pb.go'
#rm 'core/trigger.pb.gw.go'
#rm 'core/trigger.swagger.json'
#$fake git mv core ${NEW}pb
#$fake sed -i s/core/${NEW}pb/g ${NEW}pb/proto.go
#$fake sed -i s/$OLD\\\\/${NEW}\\\\/g wercker.yml
#$fake sed "s/blue_print/$NEW/g" $ROOT/templates/service/blue_printpb/generate-protobuf.sh > ${NEW}pb/generate-protobuf.sh
#
#
#while (( "$#" )); do
#  replace_dir $1
#  shift
#done
#Replacing core -> triggerpb in state/store.go
#Replacing core -> triggerpb in consumer/consumer.go
#
#set +v
#termie@golfer:github.com/wercker/trigger % git status                                                                                      [pbswitch] 11:02:26
#On branch pbswitch
#Changes to be committed:
#  (use "git reset HEAD <file>..." to unstage)
#
#	deleted:    core/trigger.pb.go
#	deleted:    core/trigger.pb.gw.go
#	deleted:    core/trigger.swagger.json
#	renamed:    core/trigger.proto -> trigger.proto
#	renamed:    core/generate-protobuf.sh -> triggerpb/generate-protobuf.sh
#	renamed:    core/proto.go -> triggerpb/proto.go
#
#Changes not staged for commit:
#  (use "git add <file>..." to update what will be committed)
#  (use "git checkout -- <file>..." to discard changes in working directory)
#
#	modified:   consumer/consumer.go
#	modified:   state/store.go
#	modified:   triggerpb/generate-protobuf.sh
#	modified:   triggerpb/proto.go

set -e

ROOT=$(dirname $BASH_SOURCE)

OLD=$1
NEW=$2

if [[ -z "$OLD" || -z "$NEW" ]]; then
  echo "Need at least two args: migrate_pb.sh <old> <new> [<dir to find-replace> <dir to find-replace>]"
  echo "Example usage:"
  echo "  ../blueprint/migrate_pb.sh trigger trigger state"
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
