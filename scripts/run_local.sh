#!/bin/bash
MY_VARS="./my_local_dev_vars"
if [ ! -f $MY_VARS ]; then 
  echo "Please create $MY_VARS"
  echo "cp $MY_VARS.example $MY_VARS"
  echo "then edit as needed"
  exit 1
fi 

source ./${MY_VARS}
if [ "$?" -ne "0" ]; then
  echo "Error reading in ${MY_VARS}"
  exit 1
fi 

if [ -z "${BROKER_CMD}" ]; then 
  echo "Please ensure BROKER_CMD is defined in ${MY_VARS}"
  exit 1 
fi 

export KUBERNETES_SERVICE_HOST=${OPENSHIFT_SERVER_HOST}
export KUBERNETES_SERVICE_PORT=${OPENSHIFT_SERVER_PORT}

BROKER_CONFIG=${GENERATED_BROKER_CONFIG}
if [ ! -z "$1" ]; then
  BROKER_CONFIG="$1"
fi 

if [ -z "${BROKER_CONFIG}" ]; then 
  echo "Please specify a broker configuration file to run"
  exit 1
fi 

echo "Running ${BROKER_CMD} --config ${BROKER_CONFIG}"
${BROKER_CMD} --config ${BROKER_CONFIG}