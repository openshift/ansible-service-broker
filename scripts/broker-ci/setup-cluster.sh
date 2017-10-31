#!/bin/bash

set -ex

# Set to anything to indicate this is the broker's travis job
broker_travis_job=$1

function cluster-setup () {
    git clone https://github.com/fusor/catasb

    cat <<EOF > "catasb/config/my_vars.yml"
---
dockerhub_org: ansibleplaybookbundle
broker_tag: latest
broker_kind: ClusterServiceBroker
EOF

    # Multiple gates use this script. Only the broker travis gate
    # will use ./run_gate.sh.
    if [[ $broker_travis_job ]]; then
	pushd catasb/local/gate/
	./run_gate.sh
    else
	pushd catasb/local/linux/
	./run_setup_local.sh
    fi
    popd

    if [ "$?" != "0" ]; then
	echo "setup-cluster.sh failed"
	exit 1
    fi

}

echo "========== Broker CI ==========="
echo "Setting up cluster"
cluster-setup

set +e
