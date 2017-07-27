#!/bin/bash

function provision {
    oc create -f ./scripts/broker-ci/mediawiki123.yaml
    oc create -f ./scripts/broker-ci/postgresql.yaml
    ./scripts/broker-ci/wait-for-resource.sh create pod mediawiki
    ./scripts/broker-ci/wait-for-resource.sh create pod postgresql
}

function bind {
   oc create -f ./scripts/broker-ci/bind-mediawiki-postgresql.yaml || true
    ./scripts/broker-ci/wait-for-resource.sh create bindings.v1alpha1.servicecatalog.k8s.io mediawiki-postgresql-binding
}

function pickup-pod-presets {
    echo "Waiting for broker to return bind creds"
    sleep 20
    oc delete pods $(oc get pods -o name -l app=mediawiki123 -n default | head -1 | cut -f 2 -d '/') -n default
    ./scripts/broker-ci/wait-for-resource.sh create pod mediawiki
}

function verify-ci-run {
    ROUTE=$(oc get route -n default | grep mediawiki | cut -f 4 -d ' ')/index.php/Main_Page
    BIND_CHECK=$(curl ${ROUTE}| grep "div class" | cut -f 2 -d "'")
    echo "Running: curl ${ROUTE}| grep \"div class\" | cut -f 2 -d \"'\""
    if [ "${BIND_CHECK}" = "error" ]; then
	echo "MAKE CI FAILED"
    elif [ "${BIND_CHECK}" = "" ]; then
	echo "Failed to gather data from ${ROUTE}. MAKE CI FAILED"
    else
	echo "SUCCESS"
	echo "You can double check by opening http://${ROUTE} in your browser"
    fi
}

provision
bind
pickup-pod-presets
verify-ci-run
