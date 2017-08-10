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

if ${INSECURE}; then
    FLAGS+=" --insecure"
fi

asbd -c $BROKER_CONFIG $FLAGS
