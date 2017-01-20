#!/bin/bash -e

#. shared.sh

instanceUUID="688eea24-9cf9-43e3-9942-d1863b2a16af"
planUUID="560789e6-d4fc-4bdf-b227-454002d5e7c6"
serviceUUID="86aa5be4-dad0-407c-8133-2ca47ca1511a"

req="{
  \"plan_id\": \"$planUUID\",
  \"service_id\": \"$serviceUUID\",
  \"parameters\": {
    \"MYSQL_USER\": \"username\"
  }
}"

curl \
  -X PUT \
  -H 'X-Broker-API-Version: 2.9' \
  -H 'Content-Type: application/json' \
  -d "$req" \
  -v \
  http://cap.example.com:8000/v2/service_instances/$instanceUUID

  #http://localhost:8000/v2/service_instances/$instanceUUID
