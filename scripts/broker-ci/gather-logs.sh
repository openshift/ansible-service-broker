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

function pod-logs {
    log-header "pod-logs"
    kubectl get pods --all-namespaces
    log-footer "pod-logs"
}

function secret-logs {
    log-header "secrets-logs"
    kubectl get secrets
    log-footer "secrets-logs"
}

function broker-logs {
    log-header "broker-logs"
    kubectl logs $(kubectl get pods -o name -l service=asb --all-namespaces | cut -f 2 -d '/') -c asb -n ansible-service-broker
    sleep 10
    log-footer "broker-logs"
}

function instance-logs {
    log-header "instance-logs"
    kubectl get clusterserviceclasses
    kubectl get serviceinstances
    log-footer "instance-logs"
}

function catalog-logs {
    log-header "catalog-logs"
    catalog_ns=$(kubectl get ns | grep catalog | cut -f 1 -d ' ' | head -1)
    kubectl logs $(kubectl get pods -o name -l app=controller-manager --all-namespaces | cut -f 2 -d '/') -n $catalog_ns
    log-footer "catalog-logs"
}

function print-pod-errors {
    log-header "pod-errors"
    pods=$(kubectl get pods --all-namespaces --no-headers | awk '{ print $2 }')

    for pod in $pods; do
	namespace=$(kubectl get pods  --all-namespaces --no-headers | grep $pod | awk '{ print $1 }')
	status=$(kubectl get pods $pod -n $namespace --no-headers | awk '{ print $3 }')

	echo $pod
	case "${status}" in
	    "ImagePullBackOff")
		kubectl describe pod $pod -n $namespace
		;;
	    "ErrImagePull")
		kubectl describe pod $pod -n $namespace
		;;
	    "CrashLoopBackOff")
		kubectl describe pod $pod -n $namespace
		kubectl logs $pod -n $namespace
		;;
	    "Error")
		kubectl logs $pod -n $namespace
		;;
	esac
    done
    log-footer "pod-errors"
}

function print-all-logs {
    print-pod-errors
    secret-logs
    pod-logs
    instance-logs
    broker-logs
    catalog-logs
}

print-all-logs
