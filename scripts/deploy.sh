#!/bin/bash
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT=${SCRIPT_DIR}/..
TEMPLATE_DIR="${PROJECT_ROOT}/templates"

set -e

# from makefile
BROKER_IMAGE=$1
REGISTRY=$2

# override with from my_local_dev_vars
PROJECT="ansible-service-broker"
OPENSHIFT_TARGET="https://kubernetes.default"
OPENSHIFT_USER="admin"
OPENSHIFT_PASS="admin"
REGISTRY_TYPE="dockerhub"
DEV_BROKER="true"
LAUNCH_APB_ON_BIND="false"
OUTPUT_REQUEST="true"
RECOVERY="true"
#DOCKERHUB_USERNAME="CHANGEME"
#DOCKERHUB_PASSWORD="CHANGEME"
#DOCKERHUB_ORG="ansibleplaybookbundle"

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
  echo "Error reading in ${MY_VARS}"
  exit 1
fi

function validate_var {
    if [ -z ${2+x} ]
    then
        echo "${1} is unset"
        exit 1
    fi
}

# check the variables that do not have defaults
validate_var "BROKER_IMAGE" $BROKER_IMAGE
validate_var "REGISTRY" $REGISTRY
validate_var "DOCKERHUB_USERNAME" $DOCKERHUB_USERNAME
validate_var "DOCKERHUB_PASSWORD" $DOCKERHUB_PASSWORD
validate_var "DOCKERHUB_ORG" $DOCKERHUB_ORG

# configure variables to pass to template
VARS="-p BROKER_IMAGE=${BROKER_IMAGE} -p OPENSHIFT_TARGET=${OPENSHIFT_TARGET} -p OPENSHIFT_PASS=${OPENSHIFT_PASS} -p OPENSHIFT_USER=${OPENSHIFT_USER} -p DOCKERHUB_ORG=${DOCKERHUB_ORG} -p DOCKERHUB_PASS=${DOCKERHUB_PASS} -p DOCKERHUB_USER=${DOCKERHUB_USER} -p REGISTRY_TYPE=${REGISTRY_TYPE} -p REGISTRY_URL=${REGISTRY} -p DEV_BROKER=${DEV_BROKER} -p LAUNCH_APB_ON_BIND=${LAUNCH_APB_ON_BIND} -p OUTPUT_REQUEST=${OUTPUT_REQUEST} -p RECOVERY=${RECOVERY}"

# cleanup old deployment
oc delete project ${PROJECT}
oc projects | grep ${PROJECT}
while [ $? -eq 0 ]
do
  echo "Waiting for ${PROJECT} to be deleted"
  sleep 5;
  oc projects | grep ${PROJECT}
done

# deploy
oc new-project ${PROJECT}
oc process -f deploy-ansible-service-broker.template.yaml -n ${PROJECT} ${VARS}  | oc create -f -
