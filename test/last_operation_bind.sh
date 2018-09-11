#!/bin/bash -e

INSTANCE_ID="$1"
BINDING_ID="$2"
OPERATION="$3"
PLAN_UUID="$4"
SERVICE_UUID="$5"

if [ -z "$6" ]
then
      HOSTNAME='172.17.0.1'
else
      HOSTNAME=$6
fi


validate_param() {
  if [ "$1" = "" ]
  then
    echo "Usage: $0 <instance uuid> <binding uuid> <operation> <plan uuid> <service uuid>"
    exit
  fi
}

validate_param "$INSTANCE_ID"
validate_param "$BINDING_ID"
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
    "https://broker-automation-broker.$HOSTNAME.nip.io/osb/v2/service_instances/$INSTANCE_ID/service_bindings/$BINDING_ID/last_operation?operation=$OPERATION&service_id=$SERVICE_UUID&plan_id=$PLAN_UUID"
