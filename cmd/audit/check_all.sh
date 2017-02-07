#!/bin/sh

LIST="ark auth cluster-manager enable-feature envvars inspector inviter kiddie-pool oauth reporter slack-bot stepsh-proxy tracestats trigger vpp vpp-aggregator webhook"

for x in $LIST; do
  echo "./audit $x"
  echo

  ./audit $GOPATH/src/github.com/wercker/$x

  echo
  echo
done
