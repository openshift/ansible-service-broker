#!/bin/bash

source "$(dirname "${BASH_SOURCE}")/../lib/init.sh"

PROJECT=${ASB_PROJECT}

kubectl delete ns ${PROJECT}

retries=25
for r in $(seq $retries); do
    kubectl get ns ansible-service-broker | grep ansible-service-broker
    if [ "$?" -eq 1 ]; then
	break
    fi
    sleep 4
done

kubectl delete clusterrolebindings --ignore-not-found=true asb
kubectl delete pv --ignore-not-found=true etcd

# Render the Kubernetes template
"${TEMPLATE_DIR}/k8s-template.py"

kubectl create -f "${TEMPLATE_DIR}/k8s-ansible-service-broker.yaml"
