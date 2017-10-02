#!/bin/bash

BROKER_ROOT=$(dirname "${BASH_SOURCE}")/..
ARGS="${@}"

if [[ "${KUBERNETES}" ]]; then
    echo "Using Kubernetes"
    "${BROKER_ROOT}/scripts/kubernetes/deploy.sh" $ARGS
else
    echo "Using OpenShift"
    "${BROKER_ROOT}/scripts/openshift/deploy.sh" $ARGS
fi
