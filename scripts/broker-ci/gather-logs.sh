#!/bin/bash

function travis_fold() {
  local action=$1
  local name=$2
  echo -en "travis_fold:${action}:${name}\r"
}

function log-header {
    header=$1
    travis_fold start $header
}

function log-footer {
    footer=$1
    travis_fold end $footer
}

function wait-logs {
    log-header "wait-logs"
    cat /tmp/wait-for-pods-log
    log-footer "wait-logs"
}

function pod-logs {
    log-header "pod-logs"
    oc get pods --all-namespaces
    log-footer "pod-logs"
}

function secret-logs {
    log-header "secrets-logs"
    oc get secrets --all-namespaces
    log-footer "secrets-logs"
}

function podpreset-logs {
    log-header "podpreset- logs"
    oc get podpresets --all-namespaces
    log-footer "podpreset-logs"
}

function broker-logs {
    log-header "broker-logs"
    oc logs $(oc get pods -o name -l service=asb --all-namespaces | cut -f 2 -d '/') -c asb -n ansible-service-broker
    log-footer "broker-logs"
}

function catalog-data-logs {
    log-header "catlog-data-logs"
    oc get serviceclasses --all-namespaces
    oc get instances --all-namespaces
    log-footer "catlog-data-logs"
}

function catalog-logs {
    log-header "catlog-logs"
    oc logs $(oc get pods -o name -l app=controller-manager --all-namespaces | cut -f 2 -d '/') -n service-catalog
    log-footer "catlog-logs"
}

function print-all-logs {
    wait-logs
    pod-logs
    secret-logs
    podpreset-logs
    broker-logs
    catalog-data-logs
    catalog-logs
}

print-all-logs
