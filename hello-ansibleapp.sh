#!/bin/bash
if [[ "$1" == "" ]]; then
  echo "Must provide ansibleapp action"
  exit
fi

docker run -it \
  -e "OPENSHIFT_TARGET=cap.example.com:8443" \
  -e "OPENSHIFT_USER=openshift-dev" \
  -e "OPENSHIFT_PASS=devel" \
  -e "ANSIBLEAPP_ACTION=$1" \
  fusordevel/hello-ansibleapp
