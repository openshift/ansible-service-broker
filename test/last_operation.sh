#!/bin/bash -e

#instanceUUID="688eea24-9cf9-43e3-9942-d1863b2a16af"
#planUUID="560789e6-d4fc-4bdf-b227-454002d5e7c6"
#serviceUUID="f819ddd5-37d1-4698-b3fb-a3cc99a35d2e" # ghost-ansibleapp
instanceUUID="688eea24-9cf9-43e3-9942-d1863b2a16af"
planUUID="4c10ff42-be89-420a-9bab-27a9bef9aed8"
serviceUUID="84e8173e-5ced-4489-9b4d-aac7e779c47e"

operation=$1

curl \
  -X GET \
  -H 'X-Broker-API-Version: 2.9' \
  -H 'Content-Type: application/json' \
  -v \
  "http://localhost:1338/v2/service_instances/$instanceUUID/last_operation?operation=$operation&service_id=$serviceUUID&plan_id=$planUUID"
  #"http://cap.example.com:1338/v2/service_instances/$instanceUUID/last_operation?operation=$operation&service_id=$serviceUUID&plan_id=$planUUID"
