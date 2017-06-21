#!/usr/bin/env bash

USER_ID=$(id -u)
if [ ${USER_UID} != ${USER_ID} ]; then
  sed "s@${USER_NAME}:x:\${USER_ID}:@${USER_NAME}:x:${USER_ID}:@g" ${BASE_DIR}/etc/passwd.template > /etc/passwd
fi

ASB_CONF_FILE=${ASB_CONF_FILE:-/etc/ansible-service-broker/broker-config.yaml}

if [ ! -f "$ASB_CONF_FILE" ] ; then
  echo "No config file mounted to $ASB_CONF_FILE"
fi
echo "Using config file mounted to $ASB_CONF_FILE"

ansible-service-broker -c $ASB_CONF_FILE
