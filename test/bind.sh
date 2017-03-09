#!/bin/bash -e

. shared.sh

instanceUUID="66f5e191-9e81-4019-9891-b4aa5059e9a1"
req="{
  \"plan_id\": \"$planUUID\",
  \"service_id\": \"$serviceUUID\"
}"

curl \
  -X PUT \
  -H 'X-Broker-API-Version: 2.9' \
  -H 'Content-Type: application/json' \
  -d "$req" \
  -v \
  http://cap.example.com:1338/v2/service_instances/$instanceUUID/service_bindings/$bindingUUID
