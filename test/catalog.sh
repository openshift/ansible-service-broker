#!/bin/bash -e

. shared.sh

curl \
  -H 'X-Broker-API-Version: 2.9' \
  -v \
  http://localhost:1338/v2/catalog
  #http://cap.example.com:8000/v2/catalog
