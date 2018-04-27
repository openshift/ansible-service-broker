#!/usr/bin/env bash

DEBUG_PORT=$1
BROKER_NAMESPACE=$2
BROKER_DEPLOYMENT=$3
PROJECT=${BROKER_NAMESPACE:-"automation-broker"}
DEPLOYMENT_NAME=${BROKER_DEPLOYMENT:-"automation-broker"}

oc rollout pause dc/${DEPLOYMENT_NAME} -n ${PROJECT}
oc set probe dc/${DEPLOYMENT_NAME} --remove --readiness --liveness -n ${PROJECT}
oc env dc/${DEPLOYMENT_NAME} DEBUG_ENABLED=True --overwrite=true -n ${PROJECT}
oc rollout resume dc/${DEPLOYMENT_NAME} -n ${PROJECT}
oc patch svc ${DEPLOYMENT_NAME} --patch='{"spec":{"ports":[{"name":"debug", "port":'${DEBUG_PORT}',"targetPort":'${DEBUG_PORT}}']}}' -n ${PROJECT}
sleep 5
POD_NAME="$(oc get po -o jsonpath='{.items[?(@.metadata.labels.service=="'${DEPLOYMENT_NAME}'")].metadata.name}')"
until oc get po ${POD_NAME} -o jsonpath='{.status.containerStatuses[0].ready}' | grep "true" >/dev/null 2>&1 ; do sleep 1; done

oc port-forward ${POD_NAME} ${DEBUG_PORT} -n ${PROJECT}

echo "To revert these changes, run ./scripts/clean_debug_env.sh ${PROJECT} ${DEPLOYMENT_NAME}"