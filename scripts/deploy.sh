#!/bin/bash

PROJECT_ROOT=$(dirname "${BASH_SOURCE}")/..
BROKER_IMAGE=${BROKER_IMAGE:-"docker.io/ansibleplaybookbundle/origin-ansible-service-broker:latest"}
APB_NAME=${APB_NAME:-"automation-broker-apb"}
APB_IMAGE=${APB_IMAGE:-"docker.io/automationbroker/automation-broker-apb:latest"}
ACTION=${ACTION:-"provision"}

if which kubectl; then
    CMD=kubectl
else
    CMD=oc
fi

# sed magic to make it possible to reuse install.yaml to deploy and wait for the broker
ARGS="[ \"${ACTION}\", \"-e create_broker_namespace=true\", \"-e wait_for_broker=true\", \"-e broker_image=${BROKER_IMAGE}\" ]"
APB_YAML=$(sed "s%\(image:\).*%\1 ${APB_IMAGE}%; s%\(args:\).*%\1 ${ARGS}%" ${PROJECT_ROOT}/apb/install.yaml)

echo "${APB_YAML}" | ${CMD} create -f -
while true; do
    POD_STATUS=$(${CMD} get pod -n ${APB_NAME} "${APB_NAME}" -o go-template="{{ .status.phase }}")
    echo "APB Pod Status: ${POD_STATUS}"
    if [ "${POD_STATUS}" == "Running" ]; then
        break
    fi
    sleep 1
done

${CMD} logs -n ${APB_NAME} "${APB_NAME}" -f
EXIT_CODE=$(${CMD} get pod -n ${APB_NAME} "${APB_NAME}" -o go-template="{{ range .status.containerStatuses }}{{.state.terminated.exitCode}}{{ end }}")

echo "${APB_YAML}" | ${CMD} delete -f -
if [ -z "${EXIT_CODE}" ] || [ "${EXIT_CODE}" == "<no value>" ]; then
    exit 0
else
    exit ${EXIT_CODE}
fi
