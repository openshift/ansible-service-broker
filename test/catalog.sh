#!/bin/bash -e

. shared.sh

curl \
  -H 'X-Broker-API-Version: 2.9' \
  -s \
  http://localhost:1338/v2/catalog
