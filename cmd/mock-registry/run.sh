#!/bin/bash
MOCK_REG_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
$GOPATH/bin/mock-registry --appfile $MOCK_REG_DIR/ansibleapps.yaml
