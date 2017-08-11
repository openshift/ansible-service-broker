#!/bin/bash

function print-with-red {
    echo -e "${color_red}${1}${color_norm}"
}

function print-with-green {
    echo -e "${color_green}${1}${color_norm}"
}

function print-with-yellow {
    echo -e "${color_yellow}${1}${color_norm}"
}

function convert-to-red {
    set +x

    if ${BUILD_ERROR}; then
	BUILD_ERROR="${color_red}true${color_norm}"
    fi
    if ${CLUSTER_SETUP_ERROR}; then
	CLUSTER_SETUP_ERROR="${color_red}true${color_norm}"
    fi
    if ${MAKE_DEPLOY_ERROR}; then
	MAKE_DEPLOY_ERROR="${color_red}true${color_norm}"
    fi
    if ${RESOURCE_ERROR}; then
	RESOURCE_ERROR="${color_red}true${color_norm}"
    fi
    if ${BIND_ERROR}; then
	BIND_ERROR="${color_red}true${color_norm}"
    fi
    if ${PROVISION_ERROR}; then
	PROVISION_ERROR="${color_red}true${color_norm}"
    fi
    if ${POD_PRESET_ERROR}; then
	POD_PRESET_ERROR="${color_red}true${color_norm}"
    fi
    if ${VERIFY_CI_ERROR}; then
	VERIFY_CI_ERROR="${color_red}true${color_norm}"
    fi
}

function error-variables {
    set +x

    print-with-green "##### CLUSTER SETUP VARIABLE LIST #####"
    print-with-yellow "BUILD_ERROR: ${BUILD_ERROR}"
    print-with-yellow "CLUSTER_SETUP_ERROR: ${CLUSTER_SETUP_ERROR}"
    print-with-yellow "MAKE_DEPLOY_ERROR: ${MAKE_DEPLOY_ERROR}"

    print-with-green "##### GATE VARIABLE LIST #####"
    print-with-yellow "RESOURCE_ERROR: ${RESOURCE_ERROR}"
    print-with-yellow "BIND_ERROR: ${BIND_ERROR}"
    print-with-yellow "PROVISION_ERROR: ${PROVISION_ERROR}"
    print-with-yellow "POD_PRESET_ERROR: ${POD_PRESET_ERROR}"
    print-with-yellow "VERIFY_CI_ERROR: ${VERIFY_CI_ERROR}"
}

function error-check {
    set +x

    if ${RESOURCE_ERROR}; then
	print-with-red "RESOURCE ERROR reported from ${1}"
	redirect-output
	pod-logs
	broker-logs
    elif ${BIND_ERROR}; then
	print-with-red "BIND ERROR reported from ${1}"
	redirect-output
	podpreset-logs
	secret-logs
	broker-logs
	catalog-logs
    elif ${PROVISION_ERROR}; then
	print-with-red "PROVISION ERROR reported from ${1}"
	redirect-output
	pod-logs
	broker-logs
    elif ${POD_PRESET_ERROR}; then
	print-with-red "POD PRESET ERROR reported from ${1}"
	redirect-output
	pod-logs
	secret-logs
	podpreset-logs
    elif ${VERIFY_CI_ERROR}; then
	print-with-red "VERIFY CI ERROR reported from ${1}"
	redirect-output
	print-all-logs
    fi

    if ${VERIFY_CI_ERROR} || ${POD_PRESET_ERROR} || ${PROVISION_ERROR} ||
	${BIND_ERROR} || ${RESOURCE_ERROR}; then
	restore-output
	convert-to-red
	error-variables
	exit 1
    fi

    set -x
}

function env-error-check {
    set +x

    if ${BUILD_ERROR}; then
	print-with-red "BUILD ERROR reported from ${1}"
    fi

    if ${CLUSTER_SETUP_ERROR}; then
	print-with-red "CLUSTER_SETUP ERROR reported from ${1}"
    fi

    if ${RESOURCE_ERROR}; then
	MAKE_DEPLOY_ERROR=true
	RESOURCE_ERROR=false
	print-with-red "MAKE_DEPLOY ERROR reported from ${1}"
	redirect-output
	wait-logs
	# Move restore output to the final error gathering check if
	# we need logs in other checks. See the error-check function.
	restore-output
    fi

    if ${BUILD_ERROR} || ${CLUSTER_SETUP_ERROR} || ${MAKE_DEPLOY_ERROR}; then
	convert-to-red
	error-variables
	exit 1
    fi
    set -x
}
