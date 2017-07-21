#!/bin/bash

set -ex

function provision-mediawiki() {
    instanceUUID="fbd21149-07d7-4a8a-b40b-4b815110c4cc"
    planUUID="9bb6908f-cb35-4e59-bd7a-dcef343ea28f"
    serviceUUID="${mediawiki_id}"
    namespace="default"
    req="{
  \"plan_id\": \"$planUUID\",
  \"service_id\": \"$serviceUUID\",
  \"parameters\": {
      \"mediawiki_admin_pass\": \"a\",
      \"mediawiki_site_lang\": \"en\",
      \"mediawiki_site_name\": \"MediaWiki\",
      \"mediawiki_db_schema\":\"mediawiki\",
      \"mediawiki_admin_user\": \"admin\"
      },
  \"context\": {
      \"platform\": \"kubernetes\",
      \"namespace\": \"$namespace\"}
  }
}"

    curl \
	-X PUT \
	-H 'X-Broker-API-Version: 2.9' \
	-H 'Content-Type: application/json' \
	-d "$req" \
	-v \
	"http://localhost:1338/v2/service_instances/$instanceUUID?accepts_incomplete=true"
}

function provision-postgresql() {
    instanceUUID="8c9adf85-9221-4776-aa18-cde6b7acc436"
    planUUID="4eb626a5-37bf-4be8-8a65-d1715c38de07"
    serviceUUID="${postgresql_id}"
    namespace="default"
    req="{
  \"plan_id\": \"$planUUID\",
  \"service_id\": \"$serviceUUID\",
  \"parameters\": {
      \"postgresql_user\": \"admin\",
      \"postgresql_database\": \"admin\",
      \"postgresql_version\": \"9.5\"},
  \"context\": {
      \"platform\": \"kubernetes\",
      \"namespace\": \"$namespace\"}
  }
}"

    curl \
	-X PUT \
	-H 'X-Broker-API-Version: 2.9' \
	-H 'Content-Type: application/json' \
	-d "$req" \
	-v \
	"http://localhost:1338/v2/service_instances/$instanceUUID?accepts_incomplete=true"
}

function bootstrap() {
    spec_count=$(curl \
	-H 'X-Broker-API-Version: 2.9' \
	-X POST \
	-v \
	http://localhost:1338/v2/bootstrap | jq -r '.spec_count')
}

function catalog() {
    curl \
	-H 'X-Broker-API-Version: 2.9' \
	-s \
	http://localhost:1338/v2/catalog > /tmp/catalog-info

    for i in $(seq 0 $((spec_count-1))); do
	apb=$(cat /tmp/catalog-info | jq -r ".services[$i][\"name\"]")
	if [ "${apb}" = "dockerhub-$ORG-rhscl-postgresql-apb" ]; then
	    postgresql_id=$(cat /tmp/catalog-info | jq -r ".services[$i][\"id\"]")
	elif [ "${apb}" = "dockerhub-$ORG-mediawiki123-apb" ]; then
	    mediawiki_id=$(cat /tmp/catalog-info | jq -r ".services[$i][\"id\"]")
	fi
    done

}

#ORG="ansibleplaybookbundle"
ORG="rthalliseyapbs"

bootstrap
catalog
provision-mediawiki
provision-postgresql
