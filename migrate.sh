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

  echo sed -i '' -e "\"s/$search/$replace/g\"" "\"$file\""
  sed -i '' -e "s/$search/$replace/g" "$file"
}

replace_logrus() {
  found=$(grep --binary-files=without-match "logrus" "$1")
  if [ -z "$found" ]; then
    # skip anything without logrus
    return 0
  fi
  echo "$found"
  #read -rp "Press enter to continue"

  search_and_replace "$1" 'log \"github.com\/Sirupsen\/logrus\"' '\"github.com\/wercker\/pkg\/log\"'
  search_and_replace "$1" '\"github.com\/Sirupsen\/logrus\"' '\"github.com\/wercker\/pkg\/log\"'
  search_and_replace "$1" 'logrus' 'log'
  goimports -w "$1"
}

replace_2016() {
  found=$(grep --binary-files=without-match "2016" "$1")
  if [ -z "$found" ]; then
    # skip anything without logrus
    return 0
  fi
  search_and_replace "$1" "2016" "2107"
}

update_govendor_cgo() {
  white "Updating govendor step to use CGO_ENABLED...\n"
  search_and_replace "$1" "code: govendor test" "code: CGO_ENABLED=0 govendor test"
}

update_docker_tags() {
  white "Updating docker tags to include git commit...\n"
  found=$(grep --binary-files=without-match "tag: \$WERCKER_GIT_BRANCH-\$WERCKER_GIT_COMMIT,\$WERCKER_GIT_COMMIT" "$1")
  if [ -n "$found" ]; then
    # already there, skip it
    return 0
  fi
  search_and_replace "$1" "tag: \$WERCKER_GIT_BRANCH-\$WERCKER_GIT_COMMIT" "tag: \$WERCKER_GIT_BRANCH-\$WERCKER_GIT_COMMIT,\$WERCKER_GIT_COMMIT"
}

update_tee_yml() {
  white "Updating deploy yaml to use tee...\n"
  name=$(jq -r .Name < .managed.json)
  search_and_replace "$1" "cat \*\.yml > $name.yml" "cat *.yml | tee $name.yml"
}

update_artifact_output() {
  white "Updating artifact output\n"
  name=$(jq -r .Name < .managed.json)
  found=$(grep --binary-files=without-match "cp -r \"\$WERCKER_OUTPUT_DIR/$name\" \"\$WERCKER_REPORT_ARTIFACTS_DIR\"" "$1")
  if [ -n "$found" ]; then
    # already there, skip it
    return 0
  fi

  search_and_replace "$1" "-o \\\$WERCKER_OUTPUT_DIR\/$name" "-o \\\"\\\$WERCKER_OUTPUT_DIR\/$name\\\""
  sed -i '' -e "/-o \"\$WERCKER_OUTPUT_DIR\/$name\"/a\\
\ \ \ \ \ \ \ \ \ \ cp -r \"\$WERCKER_OUTPUT_DIR\/$name\" \"\$WERCKER_REPORT_ARTIFACTS_DIR\"" $1
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
  white "Renaming service-*.template.yaml to *.template.yaml...\n"
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
  white "Ensuring go dependency $1...\n"
  # NOTE(termie): govendor list is super slow
  #found=$(govendor list +vendor | grep "$1")
  found=$(grep "\"path\"" ./vendor/vendor.json | grep "$1")
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
  update_govendor_cgo wercker.yml
  update_docker_tags wercker.yml
  update_tee_yml wercker.yml
  update_artifact_output wercker.yml
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
