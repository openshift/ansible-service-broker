#!/bin/bash

function cluster::deployments {
    action="$1"
    name="$2"
    args="${@:3}"
    oc $action deploymentconfig $name $args
}

function cluster::routes {
    action="$1"
    name="$2"
    args="${@:3}"
    oc $action route $name $args
}

function cluster::process {
    template="$1"
    asb_project="$2"
    args="${@:3}"
    oc process -f $template -n $asb_project $args | oc create -n $asb_project -f -
}

function cluster::get-ca {
    kubectl get secret ${BROKER_SVC_ACCT_SECRET_NAME} -n ${ASB_PROJECT} -o jsonpath='{ .data.service-ca\.crt }'
}

function cluster::etcd-port {
    echo "80"
}
