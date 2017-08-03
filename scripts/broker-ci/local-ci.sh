#!/bin/bash

BROKER_DIR="$(dirname "${BASH_SOURCE}")/../.."
source "${BROKER_DIR}/scripts/broker-ci/logs.sh"
source "${BROKER_DIR}/scripts/broker-ci/error.sh"

RESOURCE_ERROR=false
BIND_ERROR=false
PROVISION_ERROR=false
POD_PRESET_ERROR=false
VERIFY_CI_ERROR=false

LOCAL_CI="${LOCAL_CI:-true}"

declare -r color_start="\033["
declare -r color_red="${color_start}0;31m"
declare -r color_yellow="${color_start}0;33m"
declare -r color_green="${color_start}0;32m"
declare -r color_norm="${color_start}0m"

set -x

function provision {
    oc create -f ./scripts/broker-ci/mediawiki123.yaml || PROVISION_ERROR=true
    oc create -f ./scripts/broker-ci/postgresql.yaml || PROVISION_ERROR=true
    ./scripts/broker-ci/wait-for-resource.sh create pod mediawiki >> /tmp/wait-for-pods-log 2>&1
    ./scripts/broker-ci/wait-for-resource.sh create pod postgresql >> /tmp/wait-for-pods-log 2>&1
    error-check "provision"
}

function bind {
    print-with-green "Waiting for services to be ready"
    sleep 10
    oc create -f ./scripts/broker-ci/bind-mediawiki-postgresql.yaml || BIND_ERROR=true
    ./scripts/broker-ci/wait-for-resource.sh create bindings.v1alpha1.servicecatalog.k8s.io mediawiki-postgresql-binding >> /tmp/wait-for-pods-log 2>&1
    error-check "bind"
}

function pickup-pod-presets {
    print-with-green "Waiting for broker to return bind creds"
    sleep 20
    oc delete pods $(oc get pods -o name -l app=mediawiki123 -n default | head -1 | cut -f 2 -d '/') -n default
    ./scripts/broker-ci/wait-for-resource.sh create pod mediawiki >> /tmp/wait-for-pods-log 2>&1
    error-check "pickup-pod-presets"
}

function verify-ci-run {
    ROUTE=$(oc get route -n default | grep mediawiki | cut -f 4 -d ' ')/index.php/Main_Page
    BIND_CHECK=$(curl ${ROUTE}| grep "div class" | cut -f 2 -d "'")
    print-with-yellow "Running: curl ${ROUTE}| grep \"div class\" | cut -f 2 -d \"'\""
    if [ "${BIND_CHECK}" = "error" ]; then
	VERIFY_CI_ERROR=true
    elif [ "${BIND_CHECK}" = "" ]; then
	print-with-red "Failed to gather data from ${ROUTE}"
	VERIFY_CI_ERROR=true
    else
	print-with-green "SUCCESS"
	print-with-green "You can double check by opening http://${ROUTE} in your browser"
    fi
    error-check "verify-ci-run"
}

######
# Main
######
provision
bind
pickup-pod-presets
verify-ci-run
