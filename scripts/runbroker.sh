#!/bin/bash
PROJECT_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"/..
PROJECT_PROFILE=$1

$GOPATH/bin/broker \
  --config $PROJECT_ROOT/etc/$PROJECT_PROFILE.config.yaml \
  --scripts $PROJECT_ROOT/scripts
