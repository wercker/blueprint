#!/bin/bash

# This is a script to audit a project for blueprint compliance.
# Mostly it is to figure out which projects that use blueprint are using an
# old version and need to be updated to current standards.

# Please add checks to it as things change.
#set -e
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

fail() {
  red "✗\n"
}

success() {
  green "✓\n"
}

# Check that we are not using some package
check_not_using() {
  white "Checking not using $1... "
  found=$(grep \
    --binary-files=without-match \
    --exclude-dir vendor \
    --exclude-dir \.wercker \
    --exclude-dir \.git \
    --exclude audit.sh \
    -R "$1" .)
  if [ -n "$found" ]; then
    fail
    echo "$found"
    return 1
  else
    success
  fi
}

check_has_no() {
  white "Checking has no $1... "
  found=$(ls $1 2> /dev/null || true)
  if [ -n "$found" ]; then
    fail
    echo "$found"
    return 1
  else
    success
  fi
}

check_has() {
  white "Checking for $1... "
  found=$(ls $1 2> /dev/null || true)
  if [ -z "$found" ]; then
    fail
    echo "Did not find $1"
    return 1
  else
    success
  fi
}

check_has_deps() {
  white "Checking for dependency $1... "
  found=$(govendor list +vendor | grep "$1")
  if [ -z "$found" ]; then
    fail
    echo "Did not find $1 in vendor.json"
    return 1
  else
    success
  fi
}


main() {
  (
    cd "$1" || exit 1
    check_has_no glide.*
    check_not_using "github.com/Sirupsen/logrus"
    check_not_using "github.com/codegangsta/cli"
    check_not_using 2016
    check_has "core/generate-protobuf.sh"
    check_has ".managed.json"
    check_has "version.go"
    check_has "deployment/deployment.template.yml"
    check_has_deps "github.com/wercker/pkg/log"
  )
}

# Initial values
WATCH=0

# Check some args
while [ ! $# -eq 0 ]
do
  case "$1" in
    --watch | -w)
      shift
      WATCH=$1
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

if [ ! "$WATCH" -eq 0 ]; then
  clear
  while true
  do
    tput cup 0 0
    COLS=$(tput cols)
    printf "Every %ss: audit.sh %s\n" "$WATCH" "$CHECKDIR"
    tput cup 0 $((COLS-28))
    date
    main "$CHECKDIR"
    sleep "$WATCH"
  done
else
  main "$CHECKDIR"
fi

