#!/bin/bash -e

INSTANCE_ID=$1

if [ "$INSTANCE_ID" = "" ]
then
  echo "Usage: $0 <instance uuid>"
  exit
fi

curl \
    -k \
    -X DELETE \
    -H "Authorization: bearer $(oc whoami -t)" \
    -H "Content-type: application/json" \
    -H "Accept: application/json" \
    -H "X-Broker-API-Originating-Identity: " \
    "https://asb-1338-ansible-service-broker.172.17.0.1.nip.io/ansible-service-broker/v2/service_instances/$INSTANCE_ID"
