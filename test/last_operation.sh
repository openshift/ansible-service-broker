#!/bin/bash -e

INSTANCE_ID="$1"
OPERATION="$2"
PLAN_UUID="$3"
SERVICE_UUID="$4"

validate_param() {
  if [ "$1" = "" ]
  then
    echo "Usage: $0 <instance uuid> <binding uuid> <plan uuid> <service uuid>"
    exit
  fi
}

validate_param "$INSTANCE_ID"
validate_param "$OPERATION"
validate_param "$PLAN_UUID"
validate_param "$SERVICE_UUID"

curl \
    -k \
    -X GET \
    -H "Authorization: bearer $(oc whoami -t)" \
    -H "Content-type: application/json" \
    -H "Accept: application/json" \
    -H "X-Broker-API-Originating-Identity: " \
    "https://asb-1338-ansible-service-broker.172.17.0.1.nip.io/ansible-service-broker/v2/service_instances/$INSTANCE_ID/last_operation?operation=$OPERATION&service_id=$SERVICE_UUID&plan_id=$PLAN_UUID"
