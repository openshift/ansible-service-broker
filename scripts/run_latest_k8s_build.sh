#!/bin/bash

TAG=${TAG:-"release-1.2"}
APB_NAME=${APB_NAME:-"automation-broker-apb"}
APB_IMAGE=${APB_IMAGE:-"docker.io/automationbroker/automation-broker-apb:release-1.2"}
BROKER_NAME=${BROKER_NAME:-"ansible-service-broker"}
BROKER_IMAGE="docker.io/ansibleplaybookbundle/origin-ansible-service-broker:${TAG}"
BROKER_NAMESPACE=${BROKER_NAMESPACE:-"ansible-service-broker"}
HELM=${HELM:-"false"}

function ansible-service-broker {
    if [ "$TAG" == "canary" ]; then
        make build-image TAG="${TAG}"
    fi

    kubectl create ns $BROKER_NAMESPACE
    kubectl create serviceaccount $APB_NAME --namespace $BROKER_NAMESPACE
    kubectl create clusterrolebinding $APB_NAME --clusterrole=cluster-admin --serviceaccount=$BROKER_NAMESPACE:$APB_NAME
    kubectl run $APB_NAME \
        --namespace=$BROKER_NAMESPACE \
        --image=$APB_IMAGE \
        --restart=Never \
        --attach=true \
        --serviceaccount=$APB_NAME \
        -- provision \
            -e broker_image_tag=$TAG \
            -e broker_image=$BROKER_IMAGE \
            -e broker_name=$BROKER_NAME \
            -e broker_helm_enabled=$HELM
    kubectl delete pod $APB_NAME --namespace $BROKER_NAMESPACE
    kubectl delete serviceaccount $APB_NAME --namespace $BROKER_NAMESPACE
    kubectl delete clusterrolebinding $APB_NAME
}

echo "========================================================================"
echo "                       RUN_LATEST_K8s_BUILD"
echo "========================================================================"
echo ""
echo " This script expects a running kubernetes cluster and a service-catalog."
echo ""
echo " Setup minikube: https://kubernetes.io/docs/getting-started-guides/minikube/"
echo " Setup service-catalog: https://github.com/kubernetes-incubator/service-catalog/blob/master/docs/install.md#helm"
echo ""
echo "========================================================================"
echo ""

echo "Starting the Ansible Service Broker"
ansible-service-broker
