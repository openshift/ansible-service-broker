#!/bin/bash

BROKER_DIR="$(dirname "${BASH_SOURCE}")/../.."
source "${BROKER_DIR}/scripts/broker-ci/error.sh"
source "${BROKER_DIR}/scripts/broker-ci/logs.sh"

BUILD_ERROR=false
MAKE_DEPLOY_ERROR=false
CLUSTER_SETUP_ERROR=false
RESOURCE_ERROR=false

#set -ex
set -x

function cluster-setup () {
    git clone https://github.com/rthallisey/catasb
    pushd catasb
    git fetch
    git checkout gate-debugging
    popd
#    git clone https://github.com/fusor/catasb

    cat <<EOF > "catasb/config/my_vars.yml"
---
dockerhub_user_name: changeme
dockerhub_org: ansibleplaybookbundle
dockerhub_user_password: changeme
broker_tag: latest
broker_kind: ClusterServiceBroker
EOF

    pushd catasb/local/gate/
    ./run_gate.sh || CLUSTER_SETUP_ERROR=true
    popd

    echo nsenter findmnt -o target,fstype --noheadings --first-only --target /var/lib/origin/
    apt-cache policy docker
    oc version
    lsb_release -a
    mount
    df -h -T

    #oc cluster up --routing-suffix=172.17.0.1.nip.io --public-hostname=172.17.0.1 --image=docker.io/openshift/origin --version=latest  --service-catalog=true --server-loglevel=4 &
    oc cluster up --routing-suffix=172.17.0.1.nip.io --public-hostname=172.17.0.1 --image=docker.io/openshift/origin --version=latest &
    sleep 200
    sleep 200
    sleep 200
    sleep 200

    sudo journalctl -u docker
    sudo docker ps -a

    ls /var/lib/origin/
    sudo mkdir -p ~/.kube
    sudo cp /var/lib/origin/openshift.local.config/master/admin.kubeconfig ~/.kube/config
    sudo chown $(whoami): ~/.kube/config
    oc login --insecure-skip-tls-verify 172.17.0.1:8443 -u system:admin

    oc get pods --all-namespaces

    pod_uid=$(sudo ls /var/lib/origin/openshift.local.volumes/pods/ | head -1)
    vol_path="volumes/kubernetes.io~secret/apiserver-ssl"
    nsenter findmnt -o target,fstype --noheadings --first-only --target /var/lib/origin/openshift.local.volumes/pods/$pod_uid/$vol_path

    oc describe pods $(oc get pods -n kube-service-catalog | grep controller-manager | awk '{ print $1 }') -n kube-service-catalog
    oc describe pods $(oc get pods -n kube-service-catalog | grep apiserver | awk '{ print $1 }') -n kube-service-catalog
    ls -la /data-dir
    ls -la /etc/service-catalog-ssl
    ls -la /var/run/secrets/kubernetes.io/serviceaccount
    sudo docker logs $(docker ps -a | grep origin | awk '{ print $1 }')

    exit 1

    env-error-check "cluster-setup"

    cat <<EOF > "scripts/my_local_dev_vars"
CLUSTER_HOST=172.17.0.1
CLUSTER_PORT=8443

# BROKER_IP_ADDR must be the IP address of where to reach broker
#   it should not be 127.0.0.1, needs to be an address the pods will be able to reach
BROKER_IP_ADDR=${CLUSTER_HOST}
DOCKERHUB_USER="changeme"
DOCKERHUB_PASS="changeme"
DOCKERHUB_ORG="ansibleplaybookbundle"
BOOTSTRAP_ON_STARTUP="true"
BEARER_TOKEN_FILE=""
CA_FILE=""

# Always, IfNotPresent, Never
IMAGE_PULL_POLICY="Always"
EOF
}

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
	BUILD_ERROR=true
    fi
    env-error-check "make-build-image"
    set -x
}

function make-deploy {
    make deploy
    NAMESPACE="ansible-service-broker" ./scripts/broker-ci/wait-for-resource.sh create pod asb >> /tmp/wait-for-pods-log 2>&1
    env-error-check "make-deploy"
}

function local-env() {
    oc login --insecure-skip-tls-verify 172.17.0.1:8443 -u admin -p admin
    oc project default
    make-build-image
    make-deploy
}

echo "========== Broker CI ==========="
echo "Setting up cluster"
cluster-setup

echo "Setting up local environment"
local-env

set +e
