#!/usr/bin/env bash

ASB_CONF=/etc/ansible-service-broker/config.yaml

if [[ -z "${DOCKERHUB_USER}" ]] || [[ -z "${DOCKERHUB_PASS}" ]] || [[ -z "${DOCKERHUB_ORG}" ]]; then
  echo "ERROR: \$DOCKERHUB_USER and \$DOCKERHUB_PASS environment vars must be defined!"
  echo "These are required bootstrapping ansibleapp metadata from Dockerhub"
  echo "Vars can be set with docker run -e. Ex: -e=\"DOCKERHUB_USER=eriknelson\""
  exit 1
else
  echo "Got DOCKERHUB credentials."
fi

if [[ -z "${OPENSHIFT_TARGET}" ]] || [[ -z "${OPENSHIFT_USER}" ]] || [[ -z "${OPENSHIFT_PASS}" ]]; then
  echo "ERROR: Missing target openshift cluster credentials."
  echo "OPENSHIFT_TARGET - Ex: localhost:8443"
  echo "OPENSHIFT_USER"
  echo "OPENSHIFT_PASS"
  exit 1
else
  echo "Got OPENSHIFT credentials."
fi

sed -i "s|{{DOCKERHUB_USER}}|${DOCKERHUB_USER}|" $ASB_CONF
sed -i "s|{{DOCKERHUB_PASS}}|${DOCKERHUB_PASS}|" $ASB_CONF
sed -i "s|{{DOCKERHUB_ORG}}|${DOCKERHUB_ORG}|" $ASB_CONF
sed -i "s|{{OPENSHIFT_TARGET}}|${OPENSHIFT_TARGET}|" $ASB_CONF
sed -i "s|{{OPENSHIFT_USER}}|${OPENSHIFT_USER}|" $ASB_CONF
sed -i "s|{{OPENSHIFT_PASS}}|${OPENSHIFT_PASS}|" $ASB_CONF

echo $ASB_CONF

ansible-service-broker
