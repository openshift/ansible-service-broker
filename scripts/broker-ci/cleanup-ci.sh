#!/bin/bash

oc delete -f ./templates/postgresql-mediawiki123-bind.yaml
./scripts/broker-ci/wait-for-resource.sh delete ServiceBinding mediawiki-postgresql-binding
oc delete -f ./templates/mediawiki123.yaml
oc delete -f ./templates/postgresql.yaml
oc delete dc postgresql mediawiki123 -n default
./scripts/broker-ci/wait-for-resource.sh delete pod postgresql
./scripts/broker-ci/wait-for-resource.sh delete pod mediawiki
