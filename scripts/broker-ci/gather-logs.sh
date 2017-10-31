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
    oc get dc --all-namespaces
    oc get rc --all-namespaces
    log-footer "pod-logs"
}

function secret-logs {
    log-header "secrets-logs"
    oc get secrets --all-namespaces
    log-footer "secrets-logs"
}

function broker-logs {
    log-header "broker-logs"
    oc logs $(oc get pods -o name -l service=asb --all-namespaces | cut -f 2 -d '/') -c asb -n ansible-service-broker
    sleep 10
    log-footer "broker-logs"
}

function instance-logs {
    log-header "instance-logs"
    oc get clusterserviceclasses --all-namespaces
    oc get serviceinstances --all-namespaces
    log-footer "instance-logs"
}

function catalog-logs {
    log-header "catalog-logs"
    oc logs $(oc get pods -o name -l app=controller-manager --all-namespaces | cut -f 2 -d '/') -n kube-service-catalog
    sleep 10
    log-footer "catalog-logs"
}

function print-all-logs {
    wait-logs
    pod-logs
    secret-logs
    instance-logs
    broker-logs
    catalog-logs
}

print-all-logs
