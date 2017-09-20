#!/bin/bash
source "$(dirname "${BASH_SOURCE}")/lib/init.sh"

# from makefile
BROKER_IMAGE=$1
REGISTRY=$2
DOCKERHUB_ORG=$3

# can be overridden via my_local_dev_vars
PROJECT=${ASB_PROJECT}
ASB_SCHEME="https"
ROUTING_SUFFIX="172.17.0.1.nip.io"
OPENSHIFT_TARGET="https://kubernetes.default"
REGISTRY_TYPE="dockerhub"
DEV_BROKER="true"
LAUNCH_APB_ON_BIND="false"
OUTPUT_REQUEST="true"
RECOVERY="true"
REFRESH_INTERVAL="600s"
SANDBOX_ROLE="edit"

# load development variables
asb::load_vars

# check the variables that do not have defaults
asb::validate_var "BROKER_IMAGE" $BROKER_IMAGE
asb::validate_var "REGISTRY" $REGISTRY
asb::validate_var "DOCKERHUB_USERNAME" $DOCKERHUB_USERNAME
asb::validate_var "DOCKERHUB_PASSWORD" $DOCKERHUB_PASSWORD
asb::validate_var "DOCKERHUB_ORG" $DOCKERHUB_ORG
asb::validate_var "REFRESH_INTERVAL" $REFRESH_INTERVAL

VARS+=" -p BROKER_IMAGE=${BROKER_IMAGE} \
  -p ASB_SCHEME=${ASB_SCHEME} \
  -p ROUTING_SUFFIX=${ROUTING_SUFFIX} \
  -p OPENSHIFT_TARGET=${OPENSHIFT_TARGET} \
  -p DOCKERHUB_ORG=${DOCKERHUB_ORG} \
  -p DOCKERHUB_PASS=${DOCKERHUB_PASS} \
  -p DOCKERHUB_USER=${DOCKERHUB_USER} \
  -p REGISTRY_TYPE=${REGISTRY_TYPE} \
  -p REGISTRY_URL=${REGISTRY} \
  -p DEV_BROKER=${DEV_BROKER} \
  -p LAUNCH_APB_ON_BIND=${LAUNCH_APB_ON_BIND} \
  -p OUTPUT_REQUEST=${OUTPUT_REQUEST} \
  -p RECOVERY=${RECOVERY} \
  -p REFRESH_INTERVAL=${REFRESH_INTERVAL} \
  -p SANDBOX_ROLE=${SANDBOX_ROLE}"

# cleanup old deployment
asb::delete_project ${PROJECT}

# delete the clusterrolebinding to avoid template error
oc delete clusterrolebindings --ignore-not-found=true asb

# deploy
oc new-project ${PROJECT}
oc process -f ${BROKER_TEMPLATE} \
  -n ${PROJECT} \
  ${VARS} | oc create -f -
