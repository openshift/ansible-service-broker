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
    if ${RESOURCE_ERROR}; then
	RESOURCE_ERROR="${color_red}true${color_norm}"
    fi
    if ${BIND_ERROR}; then
	BIND_ERROR="${color_red}true${color_norm}"
    fi
    if ${PROVISION_ERROR}; then
	PROVISION_ERROR="${color_red}true${color_norm}"
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
    convert-to-red

    print-with-yellow "##### ERROR VARIABLE LIST #####"
    print-with-yellow "RESOURCE_ERROR: ${RESOURCE_ERROR}"
    print-with-yellow "BIND_ERROR: ${BIND_ERROR}"
    print-with-yellow "PROVISION_ERROR: ${PROVISION_ERROR}"
    print-with-yellow "POD_PRESET_ERROR: ${POD_PRESET_ERROR}"
    print-with-yellow "VERIFY_CI_ERROR: ${VERIFY_CI_ERROR}"
    print-with-yellow "##### END ERROR VARIABLE LIST #####"
}

function error-check {
    set +x
    if ${RESOURCE_ERROR}; then
	print-with-red "RESOURCE ERROR reported from ${1}"
	error-variables
	pod-logs
	broker-logs
	exit 1
    elif ${BIND_ERROR}; then
	print-with-red "BIND ERROR reported from ${1}"
	error-variables
	podpreset-logs
	secret-logs
	broker-logs
	catalog-logs
	exit 1
    elif ${PROVISION_ERROR}; then
	print-with-red "PROVISION ERROR reported from ${1}"
	error-variables
	pod-logs
	broker-logs
	exit 1
    elif ${POD_PRESET_ERROR}; then
	print-with-red "POD PRESET ERROR reported from ${1}"
	error-variables
	pod-logs
	secret-logs
	podpreset-logs
	exit 1
    elif ${VERIFY_CI_ERROR}; then
	print-with-red "VERIFY CI ERROR reported from ${1}"
	error-variables
	print-all-logs
	exit 1
    fi
    set -x
}
