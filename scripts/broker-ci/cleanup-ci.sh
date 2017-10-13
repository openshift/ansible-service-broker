#!/bin/bash

oc delete -f ./scripts/broker-ci/bind-mediawiki-postgresql.yaml
./scripts/broker-ci/wait-for-resource.sh delete ServiceBinding mediawiki-postgresql-binding
oc delete -f ./scripts/broker-ci/mediawiki123.yaml
oc delete -f ./scripts/broker-ci/postgresql.yaml
oc delete dc postgresql mediawiki123 -n default
./scripts/broker-ci/wait-for-resource.sh delete pod postgresql
./scripts/broker-ci/wait-for-resource.sh delete pod mediawiki
