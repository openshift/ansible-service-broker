#!/bin/bash -e

INSTANCE_ID="$1"
PLAN_UUID="$2"
SERVICE_UUID="$3"

if [ -z "$4" ]
then
      HOSTNAME='172.17.0.1'
else
      HOSTNAME=$4
fi

validate_param() {
  if [ "$1" = "" ]
  then
    echo "Usage: $0 <instance uuid> <plan uuid> <service uuid>"
    exit
  fi
}

validate_param "$INSTANCE_ID"
validate_param "$PLAN_UUID"
validate_param "$SERVICE_UUID"


req="{
  \"plan_id\": \"$PLAN_UUID\",
  \"service_id\": \"$SERVICE_UUID\",
  \"context\": {
    \"platform\": \"kubernetes\",
    \"namespace\": \"blog-project\"
  },
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
    "https://broker-automation-broker.$HOSTNAME.nip.io/osb/v2/service_instances/$INSTANCE_ID?accepts_incomplete=true"
