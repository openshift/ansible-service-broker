#!/bin/bash

BROKER_ROOT="$(dirname "${BASH_SOURCE}")/.."

if [ -n "${KUBERNETES}" ]; then
    echo "Kubernetes Cluster"
    source "${BROKER_ROOT}/scripts/kubernetes/resources.sh"
else
    echo "OpenShift Cluster"
    source "${BROKER_ROOT}/scripts/openshift/resources.sh"
fi
