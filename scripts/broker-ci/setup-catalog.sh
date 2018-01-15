#!/bin/bash

# TODO: Replace with Catalog APB
function setup-helm {
    helm_version=$(curl https://github.com/kubernetes/helm/releases/latest -s -L -I -o /dev/null -w '%{url_effective}' | xargs basename)
    curl https://storage.googleapis.com/kubernetes-helm/helm-${helm_version}-linux-amd64.tar.gz -o /tmp/helm.tgz
    tar -xvf /tmp/helm.tgz

    sudo cp ./linux-amd64/helm /usr/local/bin
    sudo chmod 775 /usr/local/bin/helm
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
    --set apiserver.image="apiserver:canary" \
    --set apiserver.imagePullPolicy="Never" \
    --set controllerManager.image="controller-manager:canary" \
    --set controllerManager.imagePullPolicy="Never"
}

echo "Starting the Service Catalog"
service-catalog
