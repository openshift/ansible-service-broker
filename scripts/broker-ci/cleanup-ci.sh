#!/bin/bash

oc delete bindings.v1alpha1.servicecatalog.k8s.io mediawiki-postgresql-binding -n default
./scripts/broker-ci/wait-for-resource.sh delete bindings.v1alpha1.servicecatalog.k8s.io mediawiki-postgresql-binding
oc delete instance postgresql mediawiki -n default
oc delete dc postgresql mediawiki123 -n default
./scripts/broker-ci/wait-for-resource.sh delete pod postgresql
./scripts/broker-ci/wait-for-resource.sh delete pod mediawiki
