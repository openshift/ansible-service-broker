#!/bin/bash

set -x

for r in $(seq 100); do
    postgresql=$(oc get pods -n default | grep postgresql | grep -v deploy | awk $'{ print $3 }')
    oc get pods -n default | grep postgresql
    if [ "${postgresql}" = 'Running' ]; then
       echo "postgresql pod is running"
       break
    fi
    sleep 1
done

for r in $(seq 100); do
    mediawiki=$(oc get pods -n default | grep mediawiki | grep -v deploy | awk $'{ print $3 }')
    oc get pods -n default | grep mediawiki
    if [ "${mediawiki}" = 'Running' ]; then
       echo "mediawiki pod is running"
       break
    fi
    sleep 1
done
