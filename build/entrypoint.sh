#!/usr/bin/env bash

USER_ID=$(id -u)
if [ ${USER_UID} != ${USER_ID} ]; then
  sed "s@${USER_NAME}:x:\${USER_ID}:@${USER_NAME}:x:${USER_ID}:@g" ${BASE_DIR}/etc/passwd.template > /etc/passwd
fi

ASB_CONF=/etc/ansible-service-broker/broker-config.yaml

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
  sed -i "s/  target: {{OPENSHIFT_TARGET}}//" $ASB_CONF
  sed -i "s/  user: {{OPENSHIFT_USER}}//" $ASB_CONF
  sed -i "s/  pass: {{OPENSHIFT_PASS}}//" $ASB_CONF
else
  echo "Got OPENSHIFT credentials."
  sed -i "s|{{OPENSHIFT_TARGET}}|${OPENSHIFT_TARGET}|" $ASB_CONF
  sed -i "s|{{OPENSHIFT_USER}}|${OPENSHIFT_USER}|" $ASB_CONF
  sed -i "s|{{OPENSHIFT_PASS}}|${OPENSHIFT_PASS}|" $ASB_CONF
fi

echo $ASB_CONF

ansible-service-broker
