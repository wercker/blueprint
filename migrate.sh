#!/bin/bash

# This is a script to help make automatic changes to a project
# for blueprint compliance.

# Ideally any changes to the base blueprint code that are automatically
# updateable would happen in here.
RED='\033[0;31m'
GREEN='\033[1;32m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color

red() {
  printf "${RED}$@${NC}"
}

white() {
  printf "${WHITE}$@${NC}"
}

green() {
  printf "${GREEN}$@${NC}"
}

walk() {
  white "Walking with $1...\n"
  found=$(find . -type f | grep -v vendor | grep -v .wercker | grep -v .git)
  for x in $found
  do
    if [ "$DEBUG" ]; then
      echo "  $x"
    fi
    eval "$1 $x"
  done
}

search_and_replace() {
  file=$1
  search=$2
  replace=$3

  sed -i '' -e "s/$search/$replace/g" "$file"
}

replace_logrus() {
  f=$1
  found=$(grep --binary-files=without-match "logrus" "$f")
  if [ -z "$found" ]; then
    # skip anything without logrus
    return 0
  fi
  echo "$found"
  read -rp "Press enter to continue"

  search_and_replace "$f" 'log \"github.com\/Sirupsen\/logrus\"' '\"github.com\/wercker\/pkg\/log\"'
  goimports -w "$f"
}

replace_2016() {
  found=$(grep --binary-files=without-match "2016" "$f")
  if [ -z "$found" ]; then
    # skip anything without logrus
    return 0
  fi
  search_and_replace "$1" "2016" "2107"
}

update_managed_json() {
  white "Updating .managed.json...\n"
  # server port
  server_port=$(grep -A 1 '"port",' server.go | grep "Value:" | sed -e 's/.* \([0-9]\{4,6\}\),/\1/')
  gateway_port=$(grep -A 1 '"port",' gateway.go | grep "Value:" | sed -e 's/.* \([0-9]\{4,6\}\),/\1/')
  description=$(grep "app.Usage" main.go | sed -e 's/.*"\(.*\)"/\1/')
  name=$(basename "$PWD")
  year=2017

  cat <<EOF > .managed.json
{
  "Template": "service",
  "Name": "$name",
  "Port": $server_port,
  "Gateway": $gateway_port,
  "Year": "$year",
  "Description": "$description"
}
EOF
  git diff .managed.json
}

rename_yaml() {
  files=$(ls deployment/$name-*.yml 2> /dev/null)
  if [ -z "$files" ]; then
    return 0
  fi
  white "Renaming deployment/*.yml\n"
  name=$(jq -r .Name < .managed.json)
  for x in $files;
  do
    new_name=$(echo $x | sed -e "s/$name-//")
    if [ "$DEBUG" ]; then
      echo "$x -> $new_name"
    fi
    git mv "$x" "$new_name"
  done
}

ensure_dep() {
  white "Checking for dependency $1...\n"
  found=$(govendor list +vendor | grep "$1")
  if [ -z "$found" ]; then
    echo "Did not find $1 in vendor.json, fetching"
    govendor fetch "$1"
    return 1
  fi
}
main() {
  (
  cd "$1" || exit 1
  update_managed_json
  rename_yaml
  walk "replace_logrus"
  ensure_dep "github.com/wercker/pkg/log"
  walk "replace_2016"
  )
}

# Check some args
while [ ! $# -eq 0 ]
do
  case "$1" in
    --debug | -d)
      DEBUG=1
      ;;
    *)
      CHECKDIR=$1
      ;;
  esac
  shift
done

if [ -z "$CHECKDIR" ]; then
  CHECKDIR=.
fi

main $CHECKDIR
