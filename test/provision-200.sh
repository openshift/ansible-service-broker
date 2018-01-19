#!/bin/bash -e

INSTANCE_ID=$1
BINDING_ID=$2
PLAN_UUID="7f4a5e35e4af2beb70076e72fab0b7ff"
SERVICE_UUID="dh-postgresql-apb-s964m"

if [ "$INSTANCE_ID" = "" ]
then
  echo "Usage: $0 <instance uuid> <binding uuid>"
  exit
fi

if [ "$BINDING_ID" = "" ]
then
  echo "Usage: $0 <instance uuid> <binding uuid>"
  exit
fi

req="{
  \"plan_id\": \"$PLAN_UUID\",
  \"service_id\": \"$SERVICE_UUID\",
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
    \"openshift_user\": \"zeus\",
    \"openshift_pass\": \"zeus\",
    \"port\": \"5432\",
    \"region\": \"us-east-1\",
    \"subnet\": \"test_awsdemo_rds_group\",
    \"vpc_security_groups\": \"sg-dec9b0a1\"
  }
}"

curl \
    -k \
    -X PUT \
    -H "Authorization: bearer $(oc whoami -t)" \
    -H "Content-type: application/json" \
    -H "Accept: application/json" \
    -H "X-Broker-API-Originating-Identity: " \
    -d "$req" \
    "https://asb-1338-ansible-service-broker.172.17.0.1.nip.io/ansible-service-broker/v2/service_instances/$INSTANCE_ID?accepts_incomplete=true"


