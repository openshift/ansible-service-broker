#!/bin/bash -e

if [ -z "$1" ]
then
      HOSTNAME='172.17.0.1'
else
      HOSTNAME=$1
fi

curl \
    -k \
    -X POST \
    -H "Authorization: bearer $(oc whoami -t)" \
    -H "Content-type: application/json" \
    -H "Accept: application/json" \
    -H "X-Broker-API-Originating-Identity: " \
    -d "$req" \
    "https://broker-automation-broker.$HOSTNAME.nip.io/osb/v2/bootstrap"
