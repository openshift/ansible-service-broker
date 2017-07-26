#!/bin/bash

set -ex

function cluster-setup () {
    git clone https://github.com/fusor/catasb

    cat <<EOF > "catasb/config/my_vars.yml"
---
dockerhub_user_name: brokerciuser
dockerhub_org_name: ansibleplaybookbundle
dockerhub_user_password: brokerciuser
EOF

    pushd catasb/local/gate/
    ./run_gate.sh
    popd

    cat <<EOF > "scripts/my_local_dev_vars"
OPENSHIFT_SERVER_HOST=172.17.0.1
OPENSHIFT_SERVER_PORT=8443

# BROKER_IP_ADDR must be the IP address of where to reach broker
#   it should not be 127.0.0.1, needs to be an address the pods will be able to reach
BROKER_IP_ADDR=${OPENSHIFT_SERVER_HOST}
DOCKERHUB_USERNAME="brokerciuser"
DOCKERHUB_PASSWORD="brokerciuser"
DOCKERHUB_ORG="ansibleplaybookbundle"
BOOTSTRAP_ON_STARTUP="true"
BEARER_TOKEN_FILE=""
CA_FILE=""

# Always, IfNotPresent, Never
IMAGE_PULL_POLICY="Always"
EOF

}

echo "========== Broker CI ==========="
echo "Setting up cluster"
cluster-setup
