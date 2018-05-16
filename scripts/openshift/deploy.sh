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
SANDBOX_ROLE="edit"
BROKER_KIND="${BROKER_KIND:-ClusterServiceBroker}"
ENABLE_BASIC_AUTH=false
BROKER_CA_CERT=$(oc get secret --no-headers=true -n kube-service-catalog | grep -m 1 service-catalog-apiserver-token | oc get secret $(awk '{ print $1 }') -n kube-service-catalog -o yaml | grep service-ca.crt | awk '{ print $2 }' | cat)
TAG=${TAG:-"release-1.2"}

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
  -p SANDBOX_ROLE=${SANDBOX_ROLE} \
  -p BROKER_KIND=${BROKER_KIND} \
  -p ENABLE_BASIC_AUTH=${ENABLE_BASIC_AUTH} \
  -p BROKER_CA_CERT=${BROKER_CA_CERT} \
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
