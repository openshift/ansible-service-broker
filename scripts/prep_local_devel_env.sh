#!/bin/bash
source "$(dirname "${BASH_SOURCE}")/lib/init.sh"

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
asb::load_vars

BROKER_SVC_ACCT_NAME="asb"
BROKER_SVC_ACCT="system:serviceaccount:${ASB_PROJECT}:${BROKER_SVC_ACCT_NAME}"

# Faking out https://github.com/kubernetes/client-go/blob/master/rest/config.go#L309
export KUBERNETES_SERVICE_HOST=${CLUSTER_HOST}
export KUBERNETES_SERVICE_PORT=${CLUSTER_PORT}
if [ -n "${KUBERNETES}" ]; then
    echo "Kubernetes Cluster"
    if [ "${KUBERNETES_SERVICE_PORT}" == "8443" ]; then
	echo "ERROR: Kubernetes uses port 6443, not ${KUBERNETES_SERVICE_PORT}"
	exit 1
    fi
    source "${SCRIPT_DIR}/kubernetes/resources.sh"
else
    echo "OpenShift Cluster"
    source "${SCRIPT_DIR}/openshift/resources.sh"
fi

SVC_ACCT_TOKEN_DIR=/var/run/secrets/kubernetes.io/serviceaccount
SVC_ACCT_CA_CRT=$SVC_ACCT_TOKEN_DIR/ca.crt
SVC_ACCT_TOKEN_FILE=$SVC_ACCT_TOKEN_DIR/token
TLS_CRT=$SVC_ACCT_TOKEN_DIR/tls.crt
TLS_KEY=$SVC_ACCT_TOKEN_DIR/tls.key

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
BROKER_SVC_ACCT_SECRET_NAME=`kubectl get serviceaccount asb -n ansible-service-broker -o jsonpath='{.secrets[0].name}'`
echo "Broker Service Account Token is in secret: ${BROKER_SVC_ACCT_SECRET_NAME}"

###
# Fetch the service-ca.crt for the service account
###
SVC_ACCT_CA_CRT_DATA="$(cluster::get-ca)"
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
BROKER_SVC_ACCT_TOKEN=`kubectl get secret ${BROKER_SVC_ACCT_SECRET_NAME} -n ${ASB_PROJECT} -o jsonpath='{ .data.token }'`
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

###
# Fetch the tls.crt for the asb deployment
###
TLS_CRT_DATA=`kubectl get secret -n ${ASB_PROJECT} asb-tls -o jsonpath='{ .data.tls\.crt }'`
# Remove quotes from variable
TLS_CRT_DATA=( $(eval echo ${TLS_CRT_DATA[@]}) )
# Base64 Decode
TLS_CRT_DATA=`echo ${TLS_CRT_DATA} | base64 --decode `
if [ "$?" -ne 0 ]; then
  echo "Unable to determine tls.crt for secret asb-tls"
  exit 1
fi
echo "${TLS_CRT_DATA}" &> ${TLS_CRT}
if [ "$?" -ne "0" ]; then
  echo "Unable to write the tls.crt data for asb-tls to: ${TLS_CRT}"
  exit 1
fi
echo "TLS Cert: tls.crt"
echo -e "Wrote \n${TLS_CRT_DATA}\n to: ${TLS_CRT}\n"

###
# Fetch the tls.crt for the asb deployment
###
TLS_KEY_DATA=`kubectl get secret -n ${ASB_PROJECT} asb-tls -o jsonpath='{ .data.tls\.key }'`
# Remove quotes from variable
TLS_KEY_DATA=( $(eval echo ${TLS_KEY_DATA[@]}) )
# Base64 Decode
TLS_KEY_DATA=`echo ${TLS_KEY_DATA} | base64 --decode `
if [ "$?" -ne 0 ]; then
  echo "Unable to determine tls.crt for secret asb-tls"
  exit 1
fi
echo "${TLS_KEY_DATA}" &> ${TLS_KEY}
if [ "$?" -ne "0" ]; then
  echo "Unable to write the tls.crt data for asb-tls to: ${TLS_KEY}"
  exit 1
fi
echo "TLS Cert: tls.crt"
echo -e "Wrote \n${TLS_KEY_DATA}\n to: ${TLS_KEY}\n"
# Kill any running broker pods
cluster::deployments scale asb --replicas 0 -n ${ASB_PROJECT}
# Wait for asb pod to be destroyed
kubectl get pods -n ${ASB_PROJECT} | grep asb
while [ "$?" -ne 1 ]; do
  echo "Waiting for asb deployment to scale down"
  sleep 5
  kubectl get pods -n ${ASB_PROJECT} | grep asb
done

kubectl scale deployment etcd --replicas 0 -n ${ASB_PROJECT}
# Wait for asb pod to be destroyed
kubectl get pods -n ${ASB_PROJECT} | grep etcd
while [ "$?" -ne 1 ]; do
  echo "Waiting for etcd deployment to scale down"
  sleep 5
  kubectl get pods -n ${ASB_PROJECT} | grep etcd
done

TERMINATION="reencrypt"

kubectl delete deployment etcd -n ${ASB_PROJECT}
kubectl delete endpoints asb -n ${ASB_PROJECT}
kubectl delete endpoints asb-1338 -n ${ASB_PROJECT} --ignore-not-found
kubectl delete service asb  -n ${ASB_PROJECT}
cluster::routes delete asb-etcd -n ${ASB_PROJECT}
kubectl delete service etcd -n ${ASB_PROJECT}

# Process required changes for local development
cluster::process ${TEMPLATE_LOCAL_DEV} ${ASB_PROJECT} -p BROKER_IP_ADDR=${BROKER_IP_ADDR} -p TERMINATION=${TERMINATION}

echo "Sleeping for a few seconds to avoid issues with broker not being able to talk to etcd."
echo "Appears like there is a delay of when we create the asb-etcd route and when it is available for use"
sleep 5

if [ "$LOCAL_ETCD" == "true" ]; then
  ETCD_ROUTE="localhost"
  ETCD_PORT=2379
else
  ETCD_ROUTE=`cluster::routes get asb-etcd -n ${ASB_PROJECT} -o=jsonpath=\'\{.spec.host\}\'`
  ETCD_PORT="$(cluster::etcd-port)"
fi

echo "etcd route is at: ${ETCD_ROUTE}"
echo "etcd port is: ${ETCD_PORT}"

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
# NOTE: Careful with registry.name; it *must* match the name that was used when
# the broker was originally brought up
    name: dh
    url: https://registry.hub.docker.com
    user: ${DOCKERHUB_USERNAME}
    pass: ${DOCKERHUB_PASSWORD}
    org: ${DOCKERHUB_ORG}
    white_list:
      # will allow all the APBs to be included. You must have at least 1 white
      # list to retrieve APBs and this is the most permissive
      - ".*-apb$"
dao:
  etcd_host: ${ETCD_ROUTE}
  etcd_port: ${ETCD_PORT}
log:
  logfile: /tmp/ansible-service-broker-asb.log
  stdout: true
  level: debug
  color: true
openshift:
  host: ${CLUSTER_HOST}
  bearer_token_file:${BEARER_TOKEN_FILE}
  ca_file:${CA_FILE}
  image_pull_policy: ${IMAGE_PULL_POLICY}
  sandbox_role: ${SANDBOX_ROLE:-edit}
  namespace: ${NAMESPACE:-ansible-service-broker}
  keep_namespace: ${KEEP_NAMESPACE:-false}
  keep_namespace_on_error: ${KEEP_NAMESPACE_ON_ERROR:-true}
broker:
  dev_broker: true
  launch_apb_on_bind: false
  bootstrap_on_startup: ${BOOTSTRAP_ON_STARTUP:-true}
  recovery: true
  output_request: true
  ssl_cert_key: ${TLS_KEY}
  ssl_cert: ${TLS_CRT}
  refresh_interval: "600s"
  auto_escalate: ${AUTO_ESCALATE:-true}
EOF
