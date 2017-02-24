#!/bin/bash
oc login $OPENSHIFT_TARGET --insecure-skip-tls-verify=true -u $OPENSHIFT_USER -p $OPENSHIFT_PASS
