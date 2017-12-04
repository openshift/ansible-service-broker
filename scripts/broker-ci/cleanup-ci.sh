#!/bin/bash

oc delete -f ./templates/postgresql-mediawiki-bind.yaml
./scripts/broker-ci/wait-for-resource.sh delete ServiceBinding mediawiki-postgresql-binding
oc delete -f ./templates/mediawiki.yaml
oc delete -f ./templates/postgresql.yaml
oc delete dc postgresql mediawiki -n default
./scripts/broker-ci/wait-for-resource.sh delete pod postgresql
./scripts/broker-ci/wait-for-resource.sh delete pod mediawiki
