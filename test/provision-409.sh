#!/bin/bash -e

# same instanceUUID but different parameters
instanceUUID="abd21149-07d7-4a8a-b40b-4b815110c3cc"
planUUID="4c10ff42-be89-420a-9bab-27a9bef9aed8"
serviceUUID="0aaafc10-132b-41a8-a58c-73268ff1006a"

req="{
  \"plan_id\": \"$planUUID\",
  \"service_id\": \"$serviceUUID\",
  \"parameters\": {
    \"aws_access_key\": \"key\",
    \"aws_secret_key\": \"secret\",
    \"backup_retention\": \"0\",
    \"db_engine\": \"postgres\",
    \"db_name\": \"testdb\",
    \"db_password\": \"dbpasswd\",
    \"db_size\": \"15\",
    \"db_username\": \"dbuser\",
    \"engine_version\": \"9.6.1\",
    \"instance_type\": \"db.m3.medium\",
    \"namespace\": \"zeus-rds\",
    \"openshift_target\": \"https://172.31.6.193:8443\",
    \"openshift_user\": \"admin\",
    \"openshift_pass\": \"admin\",
    \"port\": \"5432\",
    \"region\": \"us-east-1\",
    \"subnet\": \"test_awsdemo_rds_group\",
    \"vpc_security_groups\": \"sg-dec9b0a1\"
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
