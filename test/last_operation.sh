#!/bin/bash -e

INSTANCE_ID=$1
OPERATION=$2
PLAN_UUID="7f4a5e35e4af2beb70076e72fab0b7ff"
SERVICE_UUID="dh-postgresql-apb-s964m"

if [ "$INSTANCE_ID" = "" ]
then
  echo "Usage: $0 <instance uuid> <operation uuid>"
  exit
fi

if [ "$OPERATION" = "" ]
then
  echo "Usage: $0 <instance uuid> <operation uuid>"
  exit
fi

curl \
    -k \
    -X GET \
    -H "Authorization: bearer $(oc whoami -t)" \
    -H "Content-type: application/json" \
    -H "Accept: application/json" \
    -H "X-Broker-API-Originating-Identity: " \
    "https://asb-1338-ansible-service-broker.172.17.0.1.nip.io/ansible-service-broker/v2/service_instances/$INSTANCE_ID/last_operation?operation=$OPERATION&service_id=$SERVICE_UUID&plan_id=$PLAN_UUID"
