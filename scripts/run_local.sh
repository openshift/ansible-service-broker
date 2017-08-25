#!/bin/bash
source "$(dirname "${BASH_SOURCE}")/lib/init.sh"

BROKER_CMD=${ASB_ROOT}/broker

asb::load_vars
asb::validate_var "BROKER_CMD" $BROKER_CMD
asb::validate_var "OPENSHIFT_SERVER_HOST" $OPENSHIFT_SERVER_HOST
asb::validate_var "OPENSHIFT_SERVER_PORT" $OPENSHIFT_SERVER_PORT

export KUBERNETES_SERVICE_HOST=${OPENSHIFT_SERVER_HOST}
export KUBERNETES_SERVICE_PORT=${OPENSHIFT_SERVER_PORT}

BROKER_CONFIG=$GENERATED_BROKER_CONFIG
if [ ! -z "$1" ]; then
  BROKER_CONFIG="$1"
fi

if [ -z "${BROKER_CONFIG}" ]; then
  echo "Please specify a broker configuration file to run"
  exit 1
fi

if [ "${BROKER_INSECURE}" = "true" ]; then
  echo "Running ${BROKER_CMD} --config ${BROKER_CONFIG} --insecure"
  ${BROKER_CMD} --config ${BROKER_CONFIG} --insecure
else
  echo "Running ${BROKER_CMD} --config ${BROKER_CONFIG}"
  ${BROKER_CMD} --config ${BROKER_CONFIG}
fi
