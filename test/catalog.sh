#!/bin/bash -e

. shared.sh

curl \
  -H 'X-Broker-API-Version: 2.9' \
  -v \
  http://cap.example.com:8000/v2/catalog
