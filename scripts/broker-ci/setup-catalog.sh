#!/bin/bash

# Since helm latest (v2.9) is not working in CI, pull an older version.
DESIRED_HELM_VERSION="v2.8.2"

# TODO: Replace with Catalog APB
function setup-helm {
    curl https://raw.githubusercontent.com/kubernetes/helm/master/scripts/get | DESIRED_VERSION=$DESIRED_HELM_VERSION bash
    helm init
}

function build-latest-service-catalog {
    if [ -d "/tmp/service-catalog" ]; then
	pushd /tmp/service-catalog && git pull && popd
    else
	git clone https://github.com/kubernetes-incubator/service-catalog /tmp/service-catalog
    fi

    pushd /tmp/service-catalog && make images && popd
}

function service-catalog {
    setup-helm
    NAMESPACE="kube-system" ./${BROKER_DIR}/scripts/broker-ci/wait-for-resource.sh create pod tiller

    echo "Building Latest Service Catalog Images"
    build-latest-service-catalog

    kubectl create clusterrolebinding tiller-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:default

    helm install /tmp/service-catalog/charts/catalog \
    --name catalog \
    --namespace catalog \
    --set imagePullPolicy="Never" \
    --set image="service-catalog:canary" \
    --set apiserver.verbosity="2" \
    --set controllerManager.verbosity="2"
}

echo "Starting the Service Catalog"
service-catalog
