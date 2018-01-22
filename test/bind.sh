#!/bin/bash -e

INSTANCE_ID="$1"
BINDING_ID="$2"
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
validate_param "$BINDING_ID"
validate_param "$PLAN_UUID"
validate_param "$SERVICE_UUID"


req="{
  \"plan_id\": \"$PLAN_UUID\",
  \"service_id\": \"$SERVICE_UUID\",
  \"context\": \"blog-project\",
  \"app_guid\":\"\",
  \"bind_resource\":{},
  \"parameters\": {}
}"

curl \
    -k \
    -X PUT \
    -H "Authorization: bearer $(oc whoami -t)" \
    -H "Content-type: application/json" \
    -H "Accept: application/json" \
    -H "X-Broker-API-Originating-Identity: " \
    -d "$req" \
    "https://asb-1338-ansible-service-broker.172.17.0.1.nip.io/ansible-service-broker/v2/service_instances/$INSTANCE_ID/service_bindings/$BINDING_ID?accepts_incomplete=true"


