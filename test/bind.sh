#!/bin/bash -e

. shared.sh

# mlab with params
instanceUUID="8c9adf85-9221-4776-aa18-aae7b7acc436"
req="{
  \"plan_id\": \"$planUUID\",
  \"service_id\": \"$serviceUUID\",
  \"app_guid\":\"\",
  \"bind_resource\":{},
  \"parameters\": {
    \"user\": \"acct_one\"
  }
}"

curl \
  -X PUT \
  -H 'X-Broker-API-Version: 2.9' \
  -H 'Content-Type: application/json' \
  -d "$req" \
  -v \
  http://localhost:1338/v2/service_instances/$instanceUUID/service_bindings/$bindingUUID
