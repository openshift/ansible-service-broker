#!/bin/bash

function asb::load_vars() {
    # process myvars
    MY_VARS="${SCRIPT_DIR}/my_local_dev_vars"
    if [ ! -f $MY_VARS ]; then
        echo "Please create $MY_VARS"
        echo "cp $MY_VARS.example $MY_VARS"
        echo "then edit as needed"
        exit 1
    fi

    source ${MY_VARS}
    if [ "$?" -ne "0" ]; then
        echo "Error reading in ${MY_VARS}"
        exit 1
    fi
}

function asb::validate_var() {
    if [ -z ${2+x} ]; then
        echo "${1} is unset"
        exit 1
    fi
}

function asb::delete_project() {
    PROJECT=$1

    echo "Deleting project ${PROJECT}"
    oc delete project --ignore-not-found=true "${PROJECT}"
    while [[ $(oc projects) =~ .*"${PROJECT}".* ]]; do
        echo "Waiting for ${PROJECT} to be deleted"
        sleep 5
    done
    echo "Project deleted"
}
