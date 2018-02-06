#!/bin/bash

function cluster::deployments {
    action="$1"
    name="$2"
    args="${@:3}"

    kubectl $action deployments $name $args
}

function cluster::routes {
    action="$1"
    name="$2"
    args="${@:3}"

    if [ "${action}" == "delete" ]; then
	echo "Kubernetes doesn't need to delete a route"
    elif [[ "${args}" == *"jsonpath"* ]]; then
	for r in $(seq 10); do
	    endpoint="$(kubectl get endpoints | grep etcd | awk '{ print $2 }' | cut -f 1 -d ':')"
	    if [ "${endpoint}" != "<none>" ]; then
		echo "${endpoint}"
		break
	    fi
	    sleep 1
	done
    else
	kubectl $action endpoints $name $args
    fi
}

function cluster::process {
    cat <<EOF | kubectl create -f -
apiVersion: v1
kind: Endpoints
metadata:
  name: asb
subsets:
- addresses:
  - ip: ${BROKER_IP_ADDR}
  ports:
  - name: port-1338
    port: 1338
    protocol: TCP
EOF

    kubectl create -f "${TEMPLATE_DIR}/k8s-local-dev-changes.yaml"
}

function cluster::get-ca {
    kubectl get secret ${BROKER_SVC_ACCT_SECRET_NAME} -n ${ASB_PROJECT} -o jsonpath='{ .data.ca\.crt }'
}

function cluster::etcd-port {
    kubectl get endpoints | grep etcd | awk '{ print $2 }' | cut -f 2 -d ':'
}
