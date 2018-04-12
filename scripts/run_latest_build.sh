#!/bin/bash

#
# Minimal example for deploying latest built 'Ansible Service Broker'
# on oc cluster up
#


#
# We deploy oc cluster up with an explicit hostname and routing suffix
# so that pods can access routes internally.
#
# For example, we need to register the ansible service broker route to
# the service catalog when we create the broker resource. The service
# catalog needs to be able to communicate to the ansible service broker.
#
# When we use the default "127.0.0.1.nip.io" route suffix, requests
# from inside the cluster fail with an error like:
#
#    From Service Catalog: controller manager
#    controller.go:196] Error syncing Broker ansible-service-broker:
#    Get https://asb-1338-ansible-service-broker.127.0.0.1.nip.io/v2/catalog:
#    dial tcp 127.0.0.1:443: getsockopt: connection refused
#
# To resolve this, we explicitly set the
#    --public-hostname and --routing-suffix
#
# We use the IP of the docker interface on our host for testing in a
# local environment, or the external listening IP if we want to expose
# the cluster to the outside.
#
# Below will default to grabbing the IP of docker0, typically this is
# 172.17.0.1 if not customized
#

DOCKER_IP="$(ip addr show docker0 2>/dev/null | grep -Po 'inet \K[\d.]+' 2>/dev/null)"

DOCKER_IP=${DOCKER_IP:-"127.0.0.1"}
PUBLIC_IP=${PUBLIC_IP:-$DOCKER_IP}
HOSTNAME=${PUBLIC_IP}.nip.io
ROUTING_SUFFIX="${HOSTNAME}"
ORIGIN_IMAGE=${ORIGIN_IMAGE:-"docker.io/openshift/origin"}
ORIGIN_VERSION=${ORIGIN_VERSION:-"latest"}
APB_NAME=${APB_NAME:-"automation-broker-apb"}
APB_IMAGE=${APB_IMAGE:-"docker.io/automationbroker/automation-broker-apb:latest"}
BROKER_NAME=${BROKER_NAME:-"ansible-service-broker"}
BROKER_NAMESPACE=${BROKER_NAMESPACE:-"ansible-service-broker"}

version=$(oc version | head -1)
client_version=$(echo $version | egrep -o 'v[0-9]+(\.[0-9]+)+' | tr -d v.)
if [ "$ORIGIN_VERSION" != "latest" ]; then
    origin_version=$(echo $ORIGIN_VERSION | tr -d v.)
    if (( $origin_version > $client_version )); then
        echo "WARNING: Using client version: $version with cluster version: $ORIGIN_VERSION"
    fi
fi

if (( $client_version >= 3100 )); then
    oc cluster up --image=${ORIGIN_IMAGE} \
        --tag=${ORIGIN_VERSION} \
        --enable=service-catalog,template-service-broker,router,registry,web-console \
        --routing-suffix=${ROUTING_SUFFIX} \
        --public-hostname=${HOSTNAME}
else
    oc cluster up --image=${ORIGIN_IMAGE} \
        --version=${ORIGIN_VERSION} \
        --service-catalog=true \
        --routing-suffix=${ROUTING_SUFFIX} \
        --public-hostname=${HOSTNAME}
fi

if [ "$?" -ne 0 ]; then
    echo "Error starting cluster"
    exit
fi

#
# Logging in as system:admin so we can create a clusterrolebinding and
# creating ansible-service-broker project
#
echo 'Logging in as "system:admin" to create broker resources...'
oc login -u system:admin
oc new-project $BROKER_NAMESPACE

#
# Use the automation-broker-apb to deploy the broker
oc create serviceaccount $APB_NAME --namespace $BROKER_NAMESPACE
oc create clusterrolebinding $APB_NAME --clusterrole=cluster-admin --serviceaccount=$BROKER_NAMESPACE:$APB_NAME
oc run $APB_NAME \
    --namespace=$BROKER_NAMESPACE \
    --image=$APB_IMAGE \
    --restart=Never \
    --attach=true \
    --serviceaccount=$APB_NAME \
    -- provision -e broker_name=$BROKER_NAME
if [ "$?" -ne 0 ]; then
  echo "Error deploying broker"
  exit
fi
oc delete pod -n $BROKER_NAMESPACE $APB_NAME
oc delete clusterrolebinding $APB_NAME
oc delete serviceaccount $APB_NAME

#
# Then login as 'developer'/'developer' to WebUI
# Create a project
# Deploy mediawiki to new project (use a password other than
#   admin since mediawiki forbids admin as password)
# Deploy PostgreSQL(ABP) to new project
# After they are up
# Click 'Create Binding' on the kebab menu for Mediawiki,
#   select postgres
# Click deploy on mediawiki, after it's redeployed access webui
#

echo 'NOTE: You are currently logged in as "system:admin", if you intend to use the apb tool, is is required you log in as a user with a token. "developer" is recommended.'
echo '    oc adm policy add-cluster-role-to-user cluster-admin developer'
echo '    oc login -u developer'
