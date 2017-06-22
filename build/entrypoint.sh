#!/usr/bin/env bash

USER_ID=$(id -u)
if [ ${USER_UID} != ${USER_ID} ]; then
  sed "s@${USER_NAME}:x:\${USER_ID}:@${USER_NAME}:x:${USER_ID}:@g" ${BASE_DIR}/etc/passwd.template > /etc/passwd
fi

BROKER_CONFIG=${BROKER_CONFIG:-/etc/ansible-service-broker/broker-config.yaml}

if [ ! -f "$BROKER_CONFIG" ] ; then
  echo "No config file mounted to $BROKER_CONFIG"
  exit 1
fi
echo "Using config file mounted to $BROKER_CONFIG"

ansible-service-broker -c $BROKER_CONFIG
