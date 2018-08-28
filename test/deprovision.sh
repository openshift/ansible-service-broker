#!/bin/bash -e

INSTANCE_ID=$1

if [ -z "$2" ]
then
      HOSTNAME='172.17.0.1'
else
      HOSTNAME=$2
fi

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
    "https://broker-automation-broker.$HOSTNAME.nip.io/osb/v2/service_instances/$INSTANCE_ID"
