#!/bin/bash

set -ex

function cluster-setup () {
    git clone https://github.com/fusor/catasb

    cat <<EOF > "catasb/config/my_vars.yml"
---
dockerhub_org: ansibleplaybookbundle
broker_tag: latest
broker_kind: ClusterServiceBroker
EOF

    pushd catasb/local/gate/
    ./run_gate.sh
    if [ "$?" != "0" ]; then
	echo "run_gate.sh failed"
	exit 1
    fi
    popd
}

echo "========== Broker CI ==========="
echo "Setting up cluster"
cluster-setup

set +e
