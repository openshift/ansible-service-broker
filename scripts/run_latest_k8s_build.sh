#!/bin/bash

FULL_PATH_DIR="$(dirname $(dirname $(readlink -f $0)))"
BROKER_DIR="$(dirname "${BASH_SOURCE}")/../"
TEMPLATE_DIR="${BROKER_DIR}/templates"

TAG="${TAG:-}"

function create-broker-resource {
    mkdir -p /tmp/asb-cert
    openssl req -nodes -x509 -newkey rsa:4096 -keyout /tmp/asb-cert/key.pem -out /tmp/asb-cert/cert.pem -days 365 -subj "/CN=asb.ansible-service-broker.svc"
    broker_ca_cert=$(cat /tmp/asb-cert/cert.pem | base64 -w 0)
    kubectl create secret tls asb-tls --cert="/tmp/asb-cert/cert.pem" --key="/tmp/asb-cert/key.pem" -n ansible-service-broker
    client_token=$(kubectl get secrets -n ansible-service-broker | grep client-token | awk '{ print $1}')
    broker_auth='{ "bearer": { "secretRef": { "kind": "Secret", "namespace": "ansible-service-broker", "name": "REPLACE_TOKEN_STRING" } } }'

    cat <<EOF > "/tmp/broker-resource.yaml"
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ClusterServiceBroker
metadata:
  name: ansible-service-broker
spec:
  url: "https://asb.ansible-service-broker.svc:1338/ansible-service-broker/"
  authInfo:
    ${broker_auth}
  caBundle: "${broker_ca_cert}"
EOF

    sed -i 's/REPLACE_TOKEN_STRING/'"$client_token"'/g' /tmp/broker-resource.yaml
    kubectl create -f /tmp/broker-resource.yaml -n ansible-service-broker

    broker=$(kubectl get pods -n ansible-service-broker | grep -v etcd | grep asb | awk '{ print $1}')
    NAMESPACE="ansible-service-broker" ./${BROKER_DIR}/scripts/broker-ci/wait-for-resource.sh create pod "${broker}"
}

function ansible-service-broker {
    if [ "$TAG" == "build" ]; then
	make build-image TAG="${TAG}"
	sed -i 's/origin-ansible-service-broker:latest/origin-ansible-service-broker:'"$TAG"'/g' ${TEMPLATE_DIR}/k8s-variables.yaml
    elif [ -n "$TAG" ]; then
	sed -i 's/origin-ansible-service-broker:latest/origin-ansible-service-broker:'"$TAG"'/g' ${TEMPLATE_DIR}/k8s-variables.yaml
    fi

    sed -i 's/tag: latest/tag: canary/g' ${TEMPLATE_DIR}/k8s-variables.yaml

    ./${TEMPLATE_DIR}/k8s-template.py
    kubectl create ns ansible-service-broker

    context=$(kubectl config current-context)
    cluster=$(kubectl config get-contexts $context --no-headers | awk '{ print $3 }')

    kubectl config set-context $context --cluster=$cluster --namespace=ansible-service-broker
    kubectl create -f "${TEMPLATE_DIR}/k8s-ansible-service-broker.yaml"

    create-broker-resource

    broker=$(kubectl get pods -n ansible-service-broker | grep -v etcd | grep asb | awk '{ print $1}')
    NAMESPACE="ansible-service-broker" ./${BROKER_DIR}/scripts/broker-ci/wait-for-resource.sh create pod "${broker}"
}

echo "========================================================================"
echo "                       RUN_LATEST_K8s_BUILD"
echo "========================================================================"
echo ""
echo " This script expects a running kubernetes cluster and a service-catalog."
echo ""
echo " Setup minikube: https://kubernetes.io/docs/getting-started-guides/minikube/"
echo " Setup service-catalog: https://github.com/kubernetes-incubator/service-catalog/blob/master/docs/install.md#helm"
echo ""
echo "========================================================================"
echo ""

echo "Starting the Ansible Service Broker"
ansible-service-broker
