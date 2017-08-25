#!/bin/bash

BROKER_DIR="$(dirname "${BASH_SOURCE}")/../.."
source "${BROKER_DIR}/scripts/broker-ci/logs.sh"
source "${BROKER_DIR}/scripts/broker-ci/error.sh"

BIND_ERROR=false
PROVISION_ERROR=false
POD_PRESET_ERROR=false
VERIFY_CI_ERROR=false
UNBIND_ERROR=false
DEPROVISION_ERROR=false
DEVAPI_ERROR=false

RESOURCE_ERROR="${RESOURCE_ERROR:-false}"
BUILD_ERROR="${BUILD_ERROR:-false}"
MAKE_DEPLOY_ERROR="${MAKE_DELOY_ERROR:-false}"
CLUSTER_SETUP_ERROR="${CLUSTER_SETUP_ERROR:-false}"

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

function deprovision {
    oc delete -f ./scripts/broker-ci/mediawiki123.yaml || PROVISION_ERROR=true
    oc delete -f ./scripts/broker-ci/postgresql.yaml || PROVISION_ERROR=true
    ./scripts/broker-ci/wait-for-resource.sh delete pod mediawiki >> /tmp/wait-for-pods-log 2>&1
    ./scripts/broker-ci/wait-for-resource.sh delete pod postgresql >> /tmp/wait-for-pods-log 2>&1
}

function bind {
    print-with-green "Waiting for services to be ready"
    sleep 10
    oc create -f ./scripts/broker-ci/bind-mediawiki-postgresql.yaml || BIND_ERROR=true
    ./scripts/broker-ci/wait-for-resource.sh create bindings.v1alpha1.servicecatalog.k8s.io mediawiki-postgresql-binding >> /tmp/wait-for-pods-log 2>&1
    error-check "bind"
}

function unbind {
    print-with-green "Waiting for podpresets to be removed"
    oc delete -f ./scripts/broker-ci/bind-mediawiki-postgresql.yaml || BIND_ERROR=true
    ./scripts/broker-ci/wait-for-resource.sh delete podpresets mediawiki-postgresql-binding >> /tmp/wait-for-pods-log 2>&1
}

function bind-credential-check {
    set +x
    RETRIES=10
    for x in $(seq $RETRIES); do
	oc delete pods $(oc get pods -o name -l app=mediawiki123 -n default | head -1 | cut -f 2 -d '/') -n default || BIND_ERROR=true
	./scripts/broker-ci/wait-for-resource.sh create pod mediawiki >> /tmp/wait-for-pods-log 2>&1

	# Filter for 'podpreset.admission.kubernetes.io' in the pod
	preset_test=$(oc get pods $(oc get pods -n default | grep mediawiki | awk $'{ print $1 }') -o yaml -n default | grep podpreset | awk $'{ print $1}' | cut -f 1 -d '/')
	if [ "${preset_test}" = "podpreset.admission.kubernetes.io" ]; then
	    print-with-green "Pod presets found in the MediaWiki pod"
	    break
	else
	    print-with-yellow "Pod presets not found in the MediaWiki pod"
	    print-with-yellow "Retrying..."
	fi
    done

    if [ "${x}" -eq "${RETRIES}" ]; then
	print-with-red "Pod presets aren't in the MediaWiki pod"
	BIND_ERROR=true
    fi
    set -x
}

function pickup-pod-presets {
    print-with-green "Checking if MediaWiki received bind credentials"
    bind-credential-check

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

function verify-cleanup {
  if oc get -n default podpresets mediawiki-postgresql-binding ; then
    UNBIND_ERROR=true
  elif oc get -n default dc mediawiki || oc get -n default dc postgresql ; then
    DEPROVISION_ERROR=true
  fi
}

function dev-api-test {
  print-with-green "Waiting for foo apb servicename"
  BROKERURL=$(oc get -n ansible-service-broker route -o custom-columns=host:spec.host --no-headers)
  APBID=$(curl -s -k -XPOST -u admin:admin https://$BROKERURL/apb/spec -d "apbSpec=$(base64 scripts/broker-ci/apb.yml)"| \
          python -c "import sys; import json; print json.load(sys.stdin)['services'][0]['id']")
  sleep 10
  oc delete pod -n service-catalog -l app=controller-manager

  ./scripts/broker-ci/wait-for-resource.sh create serviceclass apb-push-ansibleplaybookbundle-foo-apb >> /tmp/wait-for-pods-log 2>&1

  if ! curl -I -s -k -XDELETE  -u admin:admin https://$BROKERURL/apb/spec/$APBID | grep -q "204 No Content" ; then 
    DEVAPI_ERROR=true
  fi
}

######
# Main
######
provision
bind
pickup-pod-presets
verify-ci-run
unbind
deprovision
verify-cleanup
dev-api-test
convert-to-red
error-variables
