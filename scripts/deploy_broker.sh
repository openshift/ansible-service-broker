#!/bin/bash

#
# Helper script for deploying the Broker to an existing OpenShift cluster
# using an OpenShift template.
#

#
# VERIFY and UPDATE the variables below
#
CLUSTER_ADMIN_USER="system:admin"
TEMPLATE_URL=${TEMPLATE_URL:-"https://raw.githubusercontent.com/openshift/ansible-service-broker/master/templates/deploy-ansible-service-broker.template.yaml"}
DOCKERHUB_ORG=${DOCKERHUB_ORG:-"ansibleplaybookbundle"} # DocherHub org where APBs can be found, default 'ansibleplaybookbundle'
BROKER_IMAGE=${BROKER_IMAGE:-"ansibleplaybookbundle/origin-ansible-service-broker:latest"}
ENABLE_BASIC_AUTH=${ENABLE_BASIC_AUTH:-"false"}
PROJECT_NAME=${PROJECT_NAME:-"ansible-service-broker"}

#
# Login as the CLUSTER_ADMIN_USER
#
oc login -u ${CLUSTER_ADMIN_USER}

#
# Get the BROKER_CA_CERT
#
BROKER_CA_CERT=`oc get secret -n kube-service-catalog -o go-template='{{ range .items }}{{ if eq .type "kubernetes.io/service-account-token" }}{{ index .data "service-ca.crt" }}{{end}}{{"\n"}}{{end}}' | awk NF | tail -n 1`
if [ "${BROKER_CA_CERT}" == "" ]; then
    echo -e "\nUnable to set the BROKER_CA_CERT variable!"
    echo -e "Please VERIFY that CLUSTER_ADMIN_USER is set to a user with cluster admin privileges\n"
    exit
fi

#
# creating ${PROJECT_NAME} project
#
oc new-project ${PROJECT_NAME}

# Creating openssl certs to use.
mkdir -p /tmp/etcd-cert
openssl req -nodes -x509 -newkey rsa:4096 -keyout /tmp/etcd-cert/key.pem -out /tmp/etcd-cert/cert.pem -days 365 -subj "/CN=asb-etcd.ansible-service-broker.svc"
openssl genrsa -out /tmp/etcd-cert/MyClient1.key 2048 \
&& openssl req -new -key /tmp/etcd-cert/MyClient1.key -out /tmp/etcd-cert/MyClient1.csr -subj "/CN=client" \
&& openssl x509 -req -in /tmp/etcd-cert/MyClient1.csr -CA /tmp/etcd-cert/cert.pem -CAkey /tmp/etcd-cert/key.pem -CAcreateserial -out /tmp/etcd-cert/MyClient1.pem -days 1024

ETCD_CA_CERT=$(cat /tmp/etcd-cert/cert.pem | base64)
BROKER_CLIENT_CERT=$(cat /tmp/etcd-cert/MyClient1.pem | base64)
BROKER_CLIENT_KEY=$(cat /tmp/etcd-cert/MyClient1.key | base64)

curl -s $TEMPLATE_URL \
  | oc process \
  -n ${PROJECT_NAME} \
  -p DOCKERHUB_ORG="$DOCKERHUB_ORG" \
  -p ENABLE_BASIC_AUTH="$ENABLE_BASIC_AUTH" \
  -p ETCD_TRUSTED_CA_FILE=/var/run/etcd-auth-secret/ca.crt \
  -p BROKER_CLIENT_CERT_PATH=/var/run/asb-etcd-auth/client.crt \
  -p BROKER_CLIENT_KEY_PATH=/var/run/asb-etcd-auth/client.key \
  -p ETCD_TRUSTED_CA="$ETCD_CA_CERT" \
  -p BROKER_CLIENT_CERT="$BROKER_CLIENT_CERT" \
  -p BROKER_CLIENT_KEY="$BROKER_CLIENT_KEY" \
  -p NAMESPACE="${PROJECT_NAME}" \
  -p BROKER_URL_PREFIX="${PROJECT_NAME}" \
  -p BROKER_AUTH="{ \"bearer\": { \"secretRef\": { \"kind\": \"Secret\", \"namespace\": \"${PROJECT_NAME}\", \"name\": \"ansibleservicebroker-client\" } } }" \
  -p BROKER_IMAGE="${BROKER_IMAGE}" \
  -p BROKER_CA_CERT="$BROKER_CA_CERT" -f - | oc create -f -
if [ "$?" -ne 0 ]; then
  echo "Error processing template and creating deployment"
  exit
fi

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
