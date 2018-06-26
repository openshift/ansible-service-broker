#!/usr/bin/env bash

USER_ID=$(id -u)
if [ ${USER_UID} != ${USER_ID} ]; then
  sed "s@${USER_NAME}:x:\${USER_ID}:@${USER_NAME}:x:${USER_ID}:@g" ${BASE_DIR}/etc/passwd.template > /etc/passwd
fi

if [[ -z "$BROKER_CONFIG" ]] ; then
  echo "Broker Config environment variable not set"
  exit 1
fi

if [ ! -f "$BROKER_CONFIG" ] ; then
  echo "No config file mounted to $BROKER_CONFIG"
  exit 1
fi
echo "Using config file mounted to $BROKER_CONFIG"

if [ ${DEBUG_ENABLED:-False} == "True" ]; then
    dlv  --listen=0.0.0.0:${DEBUG_PORT} --headless=true --api-version=2 --log=true exec asbd -- -c $BROKER_CONFIG $FLAGS
else
    exec asbd -c $BROKER_CONFIG $FLAGS
fi