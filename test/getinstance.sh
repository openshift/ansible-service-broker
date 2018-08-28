#!/bin/bash -e

INSTANCE_ID="$1"

if [ -z "$2" ]
then
      HOSTNAME='172.17.0.1'
else
      HOSTNAME=$2
fi

validate_param() {
  if [ "$1" = "" ]
  then
    echo "Usage: $0 <instance uuid>"
    exit
  fi
}

validate_param "$INSTANCE_ID"

curl \
    -k \
    -X GET \
    -H "Authorization: bearer $(oc whoami -t)" \
    -H "Content-type: application/json" \
    -H "Accept: application/json" \
    -H "X-Broker-API-Originating-Identity: " \
    "https://broker-automation-broker.$HOSTNAME.nip.io/osb/v2/service_instances/$INSTANCE_ID"
