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

DOCKER_IP="$(ip addr show docker0 | grep -Po 'inet \K[\d.]+')"
PUBLIC_IP=${PUBLIC_IP:-$DOCKER_IP}
HOSTNAME=${PUBLIC_IP}.nip.io
ROUTING_SUFFIX="${HOSTNAME}"
oc cluster up --image=openshift/origin --version=v3.6.0 --service-catalog=true --routing-suffix=${ROUTING_SUFFIX} --public-hostname=${HOSTNAME}


#
# A valid dockerhub username/password is required so the broker may
# authenticate with dockerhub to:
#
#  1) inspect the available repositories in an organization
#  2) read the manifest of each repository to determine metadata about
#     the images
#
# This is how the Ansible Service Broker determines what content to
# expose to the Service Catalog
#
# Note:  dockerhub API requirements require an authenticated user only,
# the user does not need any special access beyond read access to the
# organization.
#
# By default, the Ansible Service Broker will look at the
# 'ansibleplaybookbundle' organization, this can be overridden with the
# parameter DOCKERHUB_ORG being passed into the template.
#
DOCKERHUB_USER="changeme"
DOCKERHUB_PASS="changeme"
DOCKERHUB_ORG="ansibleplaybookbundle"

#
# Disabling basic auth allows "apb push" to work.
#
ENABLE_BASIC_AUTH="false"

curl -s https://raw.githubusercontent.com/openshift/ansible-service-broker/master/templates/deploy-ansible-service-broker.template.yaml > deploy-ansible-service-broker.template.yaml

#
#  Logging in as system:admin so we can create a clusterrolebinding
#
oc login -u system:admin
oc new-project ansible-service-broker
oc process -f ./deploy-ansible-service-broker.template.yaml \
    -n ansible-service-broker \
    -p DOCKERHUB_USER="$DOCKERHUB_USER" \
    -p DOCKERHUB_PASS="$DOCKERHUB_PASS" \
    -p DOCKERHUB_ORG="$DOCKERHUB_ORG" \
    -p ENABLE_BASIC_AUTH="$ENABLE_BASIC_AUTH" | oc create -f -

if [ "$?" -ne 0 ]; then
	echo "Error processing template and creating deployment"
	exit
fi

ASB_ROUTE=`oc get routes | grep ansible-service-broker | awk '{print $2}'`

cat <<EOF | oc create -f -
    apiVersion: servicecatalog.k8s.io/v1alpha1
    kind: Broker
    metadata:
      name: ansible-service-broker
    spec:
      url: https://${ASB_ROUTE}
      authInfo:
        basicAuthSecret:
          namespace: ansible-service-broker
          name: asb-auth-secret
EOF

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
