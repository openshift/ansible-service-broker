#!/bin/bash

ACTION=$1
RESOURCE=$2
RESOURCE_NAME=$3
RESOURCE_ERROR=true
NAMESPACE="${NAMESPACE:-default}"

if [ "${RESOURCE}" = "pod" ] && [ "${ACTION}" = "create" ]; then
    for r in $(seq 100); do
	pod=$(oc get pods -n ${NAMESPACE} | grep ${RESOURCE_NAME} | awk $'{ print $3 }')
	oc get pods -n default | grep ${RESOURCE_NAME}
	if [ "${pod}" = 'Running' ]; then
	        echo "${RESOURCE_NAME} ${RESOURCE} is running"
		    RESOURCE_ERROR=false
		        break
			fi
	echo "Waiting for ${RESOURCE_NAME} ${RESOURCE} to be running"
	sleep 1
    done
elif [ "${ACTION}" = "create" ]; then
    for r in $(seq 100); do
	oc get ${RESOURCE} -n ${NAMESPACE} | grep ${RESOURCE_NAME}
	if [ $? -eq 0 ]; then
	        echo "${RESOURCE_NAME} ${RESOURCE} has been created"
		    RESOURCE_ERROR=false
		        break
			fi
	echo "Waiting for ${RESOURCE_NAME} ${RESOURCE} to be created"
	sleep 1
    done
elif [ "${ACTION}" = "delete" ]; then
    for r in $(seq 100); do
	oc get ${RESOURCE} -n ${NAMESPACE} | grep ${RESOURCE_NAME}
	if [ $? -eq 1 ]; then
	        echo "${RESOURCE_NAME} ${RESOURCE} has been deleted"
		    RESOURCE_ERROR=false
		        break
			fi
	echo "Waiting for ${RESOURCE_NAME} ${RESOURCE} to be deleted"
	sleep 1
    done
fi
