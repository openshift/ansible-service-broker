#!/usr/bin/env bash

if [[ -z "$BROKER_CONFIG" ]] ; then
  echo "Broker Config environment variable not set"
  exit 1
fi

if [ ! -f "$BROKER_CONFIG" ] ; then
  echo "No config file mounted to $BROKER_CONFIG"
  exit 1
fi
echo "Using config file mounted to $BROKER_CONFIG"

exec asbd -c $BROKER_CONFIG $FLAGS
