#!/bin/bash -e

. shared.sh

curl \
  -H 'X-Broker-API-Version: 2.9' \
  -X POST \
  -v \
  http://localhost:1338/v2/bootstrap
