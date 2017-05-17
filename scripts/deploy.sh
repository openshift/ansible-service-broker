#!/bin/bash
PROJECT_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"/..
TEMPLATE_DIR="${PROJECT_ROOT}/templates"

set -e

# Based on https://gist.github.com/pkuczynski/8665367
function parse_yaml() {
    local prefix=$2
    local s
    local w
    local fs
    s='[[:space:]]*'
    w='[a-zA-Z0-9_]*'
    fs="$(echo @|tr @ '\034')"
    sed -ne "s|^\($s\)\($w\)$s:$s\"\(.*\)\"$s\$|\1$fs\2$fs\3|p" \
        -e "s|^\($s\)\($w\)$s[:-]$s\(.*\)$s\$|\1$fs\2$fs\3|p" "$1" |
    awk -F"$fs" '{
    indent = length($1)/2;
    vname[indent] = $2;
    for (i in vname) {if (i > indent) {delete vname[i]}}
        if (length($3) > 0) {
            vn=""; for (i=0; i<indent; i++) {vn=(vn)(vname[i])("_")}
            printf("%s%s%s=(\"%s\")\n", "'"$prefix"'",vn, $2, $3);
        }
    }' | sed 's/_=/+=/g'
}

function oc_create {
    oc create -f $TEMPLATE_DIR/$@
}

parse_yaml $PROJECT_ROOT/etc/dev.config.yaml > /tmp/dev-config
sed -i "s/=(\"--\")//" /tmp/dev-config

for tpl in services.yaml route.yaml etcd-deployment.yaml broker-deployment.yaml; do
    if [ "${tpl}" == "broker-deployment.yaml" ]; then
        cp $TEMPLATE_DIR/broker-deployment_template.yaml $TEMPLATE_DIR/$tpl
        sed -i "s/{{dockerhub_pass}}/${registry_pass}/" $TEMPLATE_DIR/$tpl
        sed -i "s/{{dockerhub_user}}/${registry_user}/" $TEMPLATE_DIR/$tpl
        sed -i "s/{{dockerhub_org}}/${registry_org}/" $TEMPLATE_DIR/$tpl
        sed -i "s/{{openshift_pass}}/${openshift_pass}/" $TEMPLATE_DIR/$tpl
        sed -i "s/{{openshift_target}}/${openshift_target}/" $TEMPLATE_DIR/$tpl
        sed -i "s/{{openshift_user}}/${openshift_user}/" $TEMPLATE_DIR/$tpl
    fi
    source /tmp/dev-config
    oc_create $tpl
done
