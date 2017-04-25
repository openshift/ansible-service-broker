#!/bin/bash -e

instanceUUID="abd21149-07d7-4a8a-b40b-4b815110c3cc"
planUUID="4c10ff42-be89-420a-9bab-27a9bef9aed8"
serviceUUID="0aaafc10-132b-41a8-a58c-73268ff1006a"

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
  "http://localhost:1338/v2/service_instances/$instanceUUID?accepts_incomplete=true"
  #http://cap.example.com:1338/v2/service_instances/$instanceUUID
