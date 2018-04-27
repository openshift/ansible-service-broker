#!/usr/bin/env bash

BROKER_NAMESPACE=$1
PROJECT=${BROKER_NAMESPACE:-"automation-broker"}
BROKER_DEPLOYMENT=$2
DEPLOYMENT_NAME=${BROKER_DEPLOYMENT:-"automation-broker"}

oc rollout pause dc/${DEPLOYMENT_NAME} -n ${PROJECT}
oc env dc/${DEPLOYMENT_NAME} DEBUG_ENABLED=False --overwrite=true -n ${PROJECT}
oc set probe dc/${DEPLOYMENT_NAME} --readiness --get-url=https://:1338/healthz \
    --initial-delay-seconds=15 --success-threshold=1 --timeout-seconds=1 --period-seconds=10 --failure-threshold=3 \
    --liveness --get-url=https://:1338/healthz \
    --initial-delay-seconds=15 --success-threshold=1 --timeout-seconds=1 --period-seconds=10 --failure-threshold=3 \
    -n ${PROJECT}
oc patch -n ${PROJECT} svc ${DEPLOYMENT_NAME} --type=json -p='[
    {"op": "test",
     "path": "/spec/ports/0/name",
     "value": "debug"
    },
    {"op": "remove",
     "path": "/spec/ports/0"
    }
]'
oc rollout resume dc/${DEPLOYMENT_NAME} -n ${PROJECT}