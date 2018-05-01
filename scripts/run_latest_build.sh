#!/bin/bash

#
# Minimal example deploying the broker in a Kubernetes/OpenShift cluster
#

TAG=${TAG:-"latest"}
APB_NAME=${APB_NAME:-"automation-broker-apb"}
APB_IMAGE=${APB_IMAGE:-"docker.io/automationbroker/automation-broker-apb:latest"}
BROKER_IMAGE="docker.io/ansibleplaybookbundle/origin-ansible-service-broker:${TAG}"
BROKER_NAME=${BROKER_NAME:-"ansible-service-broker"}
BROKER_NAMESPACE=${BROKER_NAMESPACE:-"ansible-service-broker"}
HELM=${HELM:-"false"}

echo "========================================================================"
echo "                       RUN_LATEST_BUILD"
echo "========================================================================"
echo ""
echo " This script expects a running Kubernetes|OpenShift cluster and a service-catalog."
echo ""
echo " OpenShift:"
echo "   Setup OpenShift: https://github.com/openshift/origin/#installation"
echo "   You must also enable the service-catalog. This depends on the installed origin"
echo "   client version."
echo "     Origin client version < 3.10 \`--service-catalog=true\`."
echo "     Origin client version >= 3.10 \`--enable=service-catalog\`."
echo ""
echo "   NOTE: When installing the broker in OpenShift, you must be an administrative"
echo "         user, (ie: \`oc login -u system:admin\`). If you intend to use the apb"
echo "         tool, it is required you log in as a user with a token. \"developer\" is recommended."
echo "             oc adm policy add-cluster-role-to-user cluster-admin developer"
echo "             oc login -u developer"
echo ""
echo " Kubernetes:"
echo " Setup minikube: https://kubernetes.io/docs/getting-started-guides/minikube/"
echo " Setup service-catalog: https://github.com/kubernetes-incubator/service-catalog/blob/master/docs/install.md#helm"
echo ""
echo "========================================================================"
echo ""

#
# Use the automation-broker-apb to deploy the broker
kubectl create namespace $BROKER_NAMESPACE
kubectl create serviceaccount $APB_NAME --namespace $BROKER_NAMESPACE
kubectl create clusterrolebinding $APB_NAME --clusterrole=cluster-admin --serviceaccount=$BROKER_NAMESPACE:$APB_NAME
kubectl run $APB_NAME \
    --namespace=$BROKER_NAMESPACE \
    --image=$APB_IMAGE \
    --restart=Never \
    --attach=true \
    --serviceaccount=$APB_NAME \
    -- provision -e broker_name=$BROKER_NAME -e broker_helm_enabled=$HELM
if [ "$?" -ne 0 ]; then
  echo "Error deploying broker"
  exit
fi
kubectl delete pod $APB_NAME --namespace $BROKER_NAMESPACE
kubectl delete serviceaccount $APB_NAME --namespace $BROKER_NAMESPACE
kubectl delete clusterrolebinding $APB_NAME
