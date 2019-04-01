#!/usr/bin/env bash

DIR=$(cd $(dirname "$0")/../../olm-catalog && pwd)
VERSION=${VERSION:-1.0.0}

NAME=${NAME:-automationbroker}
NAMEDISPLAY=${NAME:-"Automation Broker Operator"}

indent() {
  INDENT="      "
  sed "s/^/$INDENT/" | sed "s/^${INDENT}\($1\)/${INDENT:0:-2}- \1/"
}

CRD=$(cat $(ls $DIR/$VERSION/*crd.yaml) | grep -v -- "---" | indent apiVersion)
CSV=$(cat $(ls $DIR/$VERSION/*version.yaml) | indent apiVersion)
PKG=$(cat $(ls $DIR/*package.yaml) | indent packageName)

cat <<EOF | sed 's/^  *$//'
kind: ConfigMap
apiVersion: v1
metadata:
  name: $NAME
data:
  customResourceDefinitions: |-
$CRD
  clusterServiceVersions: |-
$CSV
  packages: |-
$PKG
---
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: $NAME
spec:
  configMap: $NAME
  displayName: $NAMEDISPLAY
  publisher: Red Hat
  sourceType: internal
EOF
