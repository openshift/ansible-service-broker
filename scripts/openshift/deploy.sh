#!/bin/bash
source "$(dirname "${BASH_SOURCE}")/../lib/init.sh"

# from makefile
BROKER_IMAGE=$1
REGISTRY=$2
DOCKERHUB_ORG=$3

# can be overridden via my_local_dev_vars
PROJECT=${ASB_PROJECT}
ROUTING_SUFFIX="172.17.0.1.nip.io"
OPENSHIFT_TARGET="https://kubernetes.default"
REGISTRY_TYPE="dockerhub"
DEV_BROKER="true"
LAUNCH_APB_ON_BIND="false"
OUTPUT_REQUEST="true"
RECOVERY="true"
REFRESH_INTERVAL="600s"
FORCE_DELETE="false"
SANDBOX_ROLE="edit"
BROKER_KIND="${BROKER_KIND:-ClusterServiceBroker}"
ETCD_TRUSTED_CA_FILE="/var/run/etcd-auth-secret/ca.crt"
BROKER_CLIENT_CERT_PATH="/var/run/asb-etcd-auth/client.crt"
BROKER_CLIENT_KEY_PATH="/var/run/asb-etcd-auth/client.key"
ENABLE_BASIC_AUTH=false
BROKER_CA_CERT=$(oc get secret --no-headers=true -n kube-service-catalog | grep -m 1 service-catalog-apiserver-token | oc get secret $(awk '{ print $1 }') -n kube-service-catalog -o yaml | grep service-ca.crt | awk '{ print $2 }' | cat)
TAG="${TAG:-latest}"

#Create Certs for etcd
mkdir -p /tmp/etcd-cert
openssl req -nodes -x509 -newkey rsa:4096 -keyout /tmp/etcd-cert/key.pem -out /tmp/etcd-cert/cert.pem -days 365 -subj "/CN=asb-etcd.ansible-service-broker.svc"
openssl genrsa -out /tmp/etcd-cert/MyClient1.key 2048 \
&& openssl req -new -key /tmp/etcd-cert/MyClient1.key -out /tmp/etcd-cert/MyClient1.csr -subj "/CN=client" \
&& openssl x509 -req -in /tmp/etcd-cert/MyClient1.csr -CA /tmp/etcd-cert/cert.pem -CAkey /tmp/etcd-cert/key.pem -CAcreateserial -out /tmp/etcd-cert/MyClient1.pem -days 1024

ETCD_CA_CERT=$(cat /tmp/etcd-cert/cert.pem | base64 | tr -d " \t\n\r")
ETCD_BROKER_CLIENT_CERT=$(cat /tmp/etcd-cert/MyClient1.pem | base64 | tr -d " \t\n\r")
ETCD_BROKER_CLIENT_KEY=$(cat /tmp/etcd-cert/MyClient1.key | base64 | tr -d " \t\n\r")

# load development variables
asb::load_vars

# check the variables that do not have defaults
asb::validate_var "BROKER_IMAGE" $BROKER_IMAGE
asb::validate_var "REGISTRY" $REGISTRY
asb::validate_var "DOCKERHUB_ORG" $DOCKERHUB_ORG
asb::validate_var "REFRESH_INTERVAL" $REFRESH_INTERVAL

VARS="-p BROKER_IMAGE=${BROKER_IMAGE} \
  -p ROUTING_SUFFIX=${ROUTING_SUFFIX} \
  -p OPENSHIFT_TARGET=${OPENSHIFT_TARGET} \
  -p DOCKERHUB_ORG=${DOCKERHUB_ORG} \
  -p REGISTRY_TYPE=${REGISTRY_TYPE} \
  -p REGISTRY_URL=${REGISTRY} \
  -p DEV_BROKER=${DEV_BROKER} \
  -p LAUNCH_APB_ON_BIND=${LAUNCH_APB_ON_BIND} \
  -p OUTPUT_REQUEST=${OUTPUT_REQUEST} \
  -p RECOVERY=${RECOVERY} \
  -p REFRESH_INTERVAL=${REFRESH_INTERVAL} \
  -p FORCE_DELETE=${FORCE_DELETE} \
  -p SANDBOX_ROLE=${SANDBOX_ROLE} \
  -p BROKER_KIND=${BROKER_KIND} \
  -p ENABLE_BASIC_AUTH=${ENABLE_BASIC_AUTH} \
  -p BROKER_CA_CERT=${BROKER_CA_CERT} \
  -p ETCD_TRUSTED_CA_FILE=${ETCD_TRUSTED_CA_FILE} \
  -p BROKER_CLIENT_CERT_PATH=${BROKER_CLIENT_CERT_PATH} \
  -p BROKER_CLIENT_KEY_PATH=${BROKER_CLIENT_KEY_PATH} \
  -p ETCD_TRUSTED_CA=${ETCD_CA_CERT} \
  -p BROKER_CLIENT_CERT=${ETCD_BROKER_CLIENT_CERT} \
  -p BROKER_CLIENT_KEY=${ETCD_BROKER_CLIENT_KEY} \
  -p TAG=${TAG}"

# cleanup old deployment
asb::delete_project ${PROJECT}

# delete the broker
oc delete "${BROKER_KIND}" --ignore-not-found=true ansible-service-broker

# delete the clusterrolebinding to avoid template error
oc delete clusterrolebindings --ignore-not-found=true asb

# deploy
oc new-project ${PROJECT}
oc process -f ${BROKER_TEMPLATE} \
  -n ${PROJECT} \
  ${VARS} | oc create -f -
