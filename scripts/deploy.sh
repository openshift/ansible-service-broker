#!/bin/bash
PROJECT_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"/..
TEMPLATE_DIR="${PROJECT_ROOT}/templates"

set -e

function oc_create {
    oc create -f $TEMPLATE_DIR/$@
}

for tpl in services.yaml route.yaml etcd-deployment.yaml broker-deployment.yaml; do
    oc_create $tpl
done
