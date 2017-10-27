#!/bin/bash

BROKER_DIR="$(dirname "${BASH_SOURCE}")/../.."
source "${BROKER_DIR}/scripts/broker-ci/error.sh"

set -ex

function make-build-image {
    set +x
    RETRIES=3
    for x in $(seq $RETRIES); do
	make build-image
	if [ $? -eq 0 ]; then
	    print-with-green "Broker container completed building."
	    break
	else
	    print-with-yellow "Broker container failed to build."
	    print-with-yellow "Retrying..."
	fi
    done
    if [ "${x}" -eq "${RETRIES}" ]; then
	print-with-red "Broker container failed to build."
	exit 1
    fi

    set -x
}

function make-deploy {
    make deploy
    NAMESPACE="ansible-service-broker" ./scripts/broker-ci/wait-for-resource.sh create pod asb >> /tmp/wait-for-pods-log 2>&1
}

function local-env() {
    cat <<EOF > "scripts/my_local_dev_vars"
CLUSTER_HOST=172.17.0.1
CLUSTER_PORT=8443

# BROKER_IP_ADDR must be the IP address of where to reach broker
#   it should not be 127.0.0.1, needs to be an address the pods will be able to reach
BROKER_IP_ADDR=${CLUSTER_HOST}
DOCKERHUB_ORG="ansibleplaybookbundle"
BOOTSTRAP_ON_STARTUP="true"
BEARER_TOKEN_FILE=""
CA_FILE=""

# Always, IfNotPresent, Never
IMAGE_PULL_POLICY="Always"
EOF

    oc login --insecure-skip-tls-verify 172.17.0.1:8443 -u admin -p admin
    oc project default
    make-build-image
    make-deploy
}

echo "Setting up local environment"
local-env

set +e
