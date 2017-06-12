#!/usr/bin/env bash

ASB_CONF=/etc/ansible-service-broker/config.yaml

if [[ "${DEBUG}" = "true" ]] ; then
  echo "\n$ASB_CONF before replacements:"
  cat $ASB_CONF
fi

if [[ -z "${DOCKERHUB_USER}" ]] || [[ -z "${DOCKERHUB_PASS}" ]] || [[ -z "${DOCKERHUB_ORG}" ]]; then
  echo "ERROR: \$DOCKERHUB_USER and \$DOCKERHUB_PASS environment vars must be defined!"
  echo "These are required bootstrapping ansibleapp metadata from Dockerhub"
  echo "Vars can be set with docker run -e. Ex: -e=\"DOCKERHUB_USER=eriknelson\""
  exit 1
else
  echo "Got DOCKERHUB credentials."
  sed -i "s|{{DOCKERHUB_USER}}|${DOCKERHUB_USER}|" $ASB_CONF
  sed -i "s|{{DOCKERHUB_PASS}}|${DOCKERHUB_PASS}|" $ASB_CONF
  sed -i "s|{{DOCKERHUB_ORG}}|${DOCKERHUB_ORG}|" $ASB_CONF
fi

if [[ -z "${OPENSHIFT_TARGET}" ]] || [[ -z "${OPENSHIFT_USER}" ]] || [[ -z "${OPENSHIFT_PASS}" ]]; then
  echo "Openshift cluster credentials not provided. Assuming the broker is running inside an Openshift cluster"
  sed -i "s/openshift:.*/openshift: {}/" $ASB_CONF
  sed -i "s/\s*target:\s*{{OPENSHIFT_TARGET}}//" $ASB_CONF
  sed -i "s/\s*user:\s*{{OPENSHIFT_USER}}//" $ASB_CONF
  sed -i "s/\s*pass:\s*{{OPENSHIFT_PASS}}//" $ASB_CONF
else
  echo "Got OPENSHIFT credentials."
  sed -i "s|{{OPENSHIFT_TARGET}}|${OPENSHIFT_TARGET}|" $ASB_CONF
  sed -i "s|{{OPENSHIFT_USER}}|${OPENSHIFT_USER}|" $ASB_CONF
  sed -i "s|{{OPENSHIFT_PASS}}|${OPENSHIFT_PASS}|" $ASB_CONF
fi

echo $ASB_CONF

if [[ "${DEBUG}" = "true" ]] ; then
  echo "\n$ASB_CONF after replacements:"
  cat $ASB_CONF
fi

ansible-service-broker
