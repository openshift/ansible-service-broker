#!/bin/bash -e

curl \
    -k \
    -X GET \
    -H "Authorization: bearer $(oc whoami -t)" \
    -H "Content-type: application/json" \
    -H "Accept: application/json" \
    "https://asb-1338-ansible-service-broker.172.17.0.1.nip.io/ansible-service-broker/v2/catalog"
