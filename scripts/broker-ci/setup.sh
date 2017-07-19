#!/bin/bash

set -ex

CATASB_ROOT=$(dirname "${BASH_SOURCE}")/../../catasb

function cluster-setup (){
    git clone https://github.com/rthallisey/catasb
    cat <<EOF > "${CATASB_ROOT}/config/my_vars.yml"
---
dockerhub_user_name: brokerciuser
dockerhub_org_name: ansibleplaybookbundle
dockerhub_user_password: brokerciuser
EOF

    pushd ${CATASB_ROOT}/local/linux
    git checkout gate-testing
    ./run_setup_local.sh
    popd
}

echo "========== Broker CI ==========="
echo "Setting up cluster"
cluster-setup
