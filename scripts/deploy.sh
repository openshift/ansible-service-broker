#!/bin/bash

BROKER_ROOT=$(dirname "${BASH_SOURCE}")/..

if [[ "${KUBERNETES}" ]]; then
    echo "Using Kubernetes"
    "${BROKER_ROOT}/scripts/kubernetes/deploy.sh"
else
    echo "Using OpenShift"
    "${BROKER_ROOT}/scripts/openshift/deploy.sh"
fi
