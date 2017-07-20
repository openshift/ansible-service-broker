#!/bin/bash

set -ex

CATASB_ROOT=$(dirname "${BASH_SOURCE}")/../../catasb

function old-cluster-setup (){
    git clone https://github.com/rthallisey/catasb
    cat <<EOF > "${CATASB_ROOT}/config/my_vars.yml"
---
dockerhub_user_name: brokerciuser
dockerhub_org_name: ansibleplaybookbundle
dockerhub_user_password: brokerciuser
EOF

    pushd ${CATASB_ROOT}/local/linux
    git checkout gate-testing
    ./run_setup_local.sh
    popd
}

function cluster-setup () {
    wget https://storage.googleapis.com/kubernetes-release/release/v1.6.0/bin/linux/amd64/kubectl
    sudo mv kubectl /usr/bin
    sudo chmod 755 /usr/bin/kubectl

    wget https://mirror.openshift.com/pub/openshift-v3/clients/3.6.153/linux/oc.tar.gz
    tar -xzf oc.tar.gz -C /tmp/
    sudo mv /tmp/oc /usr/bin
    sudo chmod 755 /usr/bin/oc

    oc cluster up --image=docker.io/openshift/origin --version=v3.6.0-rc.0  --service-catalog=true
    oc login -u system:admin
    oc get pods --all-namespaces
}

echo "========== Broker CI ==========="
echo "Setting up cluster"
cluster-setup

echo "Build broker image"
make build-image

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

echo "Deploygin broker"
make deploy
