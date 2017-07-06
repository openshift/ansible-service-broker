#!/bin/bash
###
# This script is intended to allow us to run the broker locally but
# fake out the environment so it seems like it is running inside the cluster
#
# To run the broker locally we address the below isses:
# - Service Catalog needs to talk to route and have it reach the local broker
#   - Update the asb service & endpoint to point to our local broker
# - Create a route for etcd so local broker can talk to etcd
# - Generate a configuration file for broker to use
# - Set KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT to reach cluster
# - Get existing secret for the asb service account and store ca cert/token
#
###
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT=${SCRIPT_DIR}/..
TEMPLATE_DIR="${PROJECT_ROOT}/templates"

# sane defaults, can be overridden in my_vars
GENERATED_BROKER_CONFIG=${PROJECT_ROOT}/etc/generated_local_development.yaml

# process myvars
MY_VARS="${SCRIPT_DIR}/my_local_dev_vars"
if [ ! -f $MY_VARS ]; then
  echo "Please create $MY_VARS"
  echo "cp $MY_VARS.example $MY_VARS"
  echo "then edit as needed"
  exit 1
fi

source ${MY_VARS}
if [ "$?" -ne "0" ]; then
  echo "Error reading in my_local_dev_var"
  exit 1
fi

TEMPLATE_LOCAL_DEV="${TEMPLATE_DIR}/deploy-local-dev-changes.yaml"
ASB_PROJECT="ansible-service-broker"
BROKER_SVC_ACCT_NAME="asb"
BROKER_SVC_ACCT="system:serviceaccount:${ASB_PROJECT}:${BROKER_SVC_ACCT_NAME}"

# Faking out https://github.com/kubernetes/client-go/blob/master/rest/config.go#L309
export KUBERNETES_SERVICE_HOST=${OPENSHIFT_SERVER_HOST}
export KUBERNETES_SERVICE_PORT=${OPENSHIFT_SERVER_PORT}
SVC_ACCT_TOKEN_DIR=/var/run/secrets/kubernetes.io/serviceaccount
SVC_ACCT_CA_CRT=$SVC_ACCT_TOKEN_DIR/ca.crt
SVC_ACCT_TOKEN_FILE=$SVC_ACCT_TOKEN_DIR/token

# We rely on jq for parsing json data from oc/kubectl
which jq &> /dev/null
if [ "$?" -ne 0 ]; then
  echo "Please ensure 'jq' is installed and in your path"
  exit 1
fi

# We will fake out the service account directory locally on the machine
# The directory is under /var/run and likely to be deleted between reboots
if [ ! -d "$SVC_ACCT_TOKEN_DIR" ]; then
  echo "Attempting to create serviceaccount directory: ${SVC_ACCT_TOKEN_DIR}"
  sudo mkdir -p ${SVC_ACCT_TOKEN_DIR}
  if [ "$?" -ne "0" ]; then
    echo "Please create serviceaccount directory with read/write permissions for your user:  ${SVC_ACCT_TOKEN_DIR}"
    exit 1
  fi
fi
sudo chown ${USER} ${SVC_ACCT_TOKEN_DIR}
if [ "$?" -ne "0" ]; then
  echo "Please chown the serviceaccount directory so your user may read/write:  ${SVC_ACCT_TOKEN_DIR}"
  exit 1
fi


# Determine the name of the secret which has the 'asb' service account info
BROKER_SVC_ACCT_SECRET_NAME=`oc get serviceaccount asb -n ansible-service-broker -o json | jq -c '.secrets[] | select(.name | contains("asb-token"))' | jq -c '.name'`
# Remove quotes from variable
BROKER_SVC_ACCT_SECRET_NAME=( $(eval echo ${BROKER_SVC_ACCT_SECRET_NAME[@]}) )
echo "Broker Service Account Token is in secret: ${BROKER_SVC_ACCT_SECRET_NAME}"

###
# Fetch the service-ca.crt for the service account
###
SVC_ACCT_CA_CRT_DATA=`oc get secret ${BROKER_SVC_ACCT_SECRET_NAME} -n ${ASB_PROJECT} -o json | jq -c '.data["service-ca.crt"]'`
# Remove quotes from variable
SVC_ACCT_CA_CRT_DATA=( $(eval echo ${SVC_ACCT_CA_CRT_DATA[@]}) )
# Base64 Decode
SVC_ACCT_CA_CRT_DATA=`echo ${SVC_ACCT_CA_CRT_DATA} | base64 --decode `
if [ "$?" -ne 0 ]; then
  echo "Unable to determine service-ca.crt for secret '${BROKER_SVC_ACCT_SECRET_NAME}'"
  exit 1
fi
echo "${SVC_ACCT_CA_CRT_DATA}" &> ${SVC_ACCT_CA_CRT}
if [ "$?" -ne "0" ]; then
  echo "Unable to write the service-ca.crt data for ${BROKER_SVC_ACCT_SECRET_NAME} to: ${SVC_ACCT_CA_CRT}"
  exit 1
fi
echo "Service Account: ca.crt"
echo -e "Wrote \n${SVC_ACCT_CA_CRT_DATA}\n to: ${SVC_ACCT_CA_CRT}\n"

###
# Fetch the token for the service account
###
if [ ! -d $SVC_ACCT_TOKEN_DIR ]; then
  echo "Please create the directory: ${SVC_ACCT_TOKEN_DIR}"
  echo "Ensure your user can write to it."
  exit 1
fi
BROKER_SVC_ACCT_TOKEN=`oc get secret ${BROKER_SVC_ACCT_SECRET_NAME} -n ${ASB_PROJECT} -o json | jq -c '.data["token"]'`
BROKER_SVC_ACCT_TOKEN=( $(eval echo ${BROKER_SVC_ACCT_TOKEN[@]}) )
BROKER_SVC_ACCT_TOKEN=`echo ${BROKER_SVC_ACCT_TOKEN} | base64 --decode`
###
# Note:
# It is important we do __not__ append the trailing newline in the token file
# Go's ioutil module will read in the newline as part of the token which breaks it...and causes confusion tracking down
###
echo -n "${BROKER_SVC_ACCT_TOKEN}" &> $SVC_ACCT_TOKEN_FILE
if [ "$?" -ne 0 ]; then
  echo "Unable to write token to $SVC_ACCT_TOKEN_FILE"
  exit 1
fi
echo "Service Account: token"
echo -e "Wrote \n${BROKER_SVC_ACCT_TOKEN}\n to: ${SVC_ACCT_TOKEN_FILE}\n"

# Kill any running broker pods
oc scale deployments asb --replicas 0 -n ${ASB_PROJECT}
# Wait for asb pod to be destroyed
oc get pods -n ${ASB_PROJECT} | grep asb
while [ "$?" -ne 1 ]; do
  echo "Waiting for asb deployment to scale down"
  sleep 5
  oc get pods -n ${ASB_PROJECT} | grep asb
done

oc delete endpoints asb -n ${ASB_PROJECT}
oc delete service asb  -n ${ASB_PROJECT}
oc delete route asb-etcd -n ${ASB_PROJECT}
# Process required changes for local development
oc process -f ${TEMPLATE_LOCAL_DEV} -n ${ASB_PROJECT} -p BROKER_IP_ADDR=${BROKER_IP_ADDR} | oc create -n ${ASB_PROJECT} -f -

echo "Sleeping for a few seconds to avoid issues with broker not being able to talk to etcd."
echo "Appears like there is a delay of when we create the asb-etcd route and when it is available for use"
sleep 5

etcd_route=`oc get route asb-etcd -n ${ASB_PROJECT} -o=jsonpath=\'\{.spec.host\}\'`
echo "etcd route is at: ${etcd_route}"

if [ -z "$DOCKERHUB_USERNAME" ]; then
  echo "Please set the environment variable DOCKERHUB_USERNAME and re-run"
  exit 1
fi
if [ -z "$DOCKERHUB_PASSWORD" ]; then
  echo "Please set the environment variable DOCKERHUB_PASSWORD and re-run"
  exit 1
fi
if [ -z "$DOCKERHUB_ORG" ]; then
  echo "Please set the environment variable DOCKERHUB_ORG and re-run"
  exit 1
fi

cat << EOF  > ${GENERATED_BROKER_CONFIG}
---
registry:
  - type: dockerhub
    name: dockerhub
    url: https://registry.hub.docker.com
    user: ${DOCKERHUB_USERNAME}
    pass: ${DOCKERHUB_PASSWORD}
    org: ${DOCKERHUB_ORG}
dao:
  etcd_host: ${etcd_route}
  etcd_port: 80
log:
  logfile: /tmp/ansible-service-broker-asb.log
  stdout: true
  level: debug
  color: true
openshift:
  host: ${OPENSHIFT_SERVER_HOST}
  bearer_token_file:${BEARER_TOKEN_FILE}
  ca_file:${CA_FILE}
  image_pull_policy: ${IMAGE_PULL_POLICY}
broker:
  dev_broker: true
  launch_apb_on_bind: false
  recovery: true
  output_request: true
EOF

