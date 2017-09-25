#!/bin/bash

# Stolen from https://github.com/openshift/origin/blob/master/hack/lib/init.sh
# to make life easier.
function asb::absolute_path() {
  local relative_path="$1"
  local absolute_path

  pushd "${relative_path}" >/dev/null
  relative_path="$( pwd )"
  if [[ -h "${relative_path}" ]]; then
    absolute_path="$( readlink "${relative_path}" )"
  else
    absolute_path="${relative_path}"
  fi
  popd >/dev/null

  echo "${absolute_path}"
}
readonly -f asb::absolute_path

init_source="$( dirname "${BASH_SOURCE}" )/../.."
ASB_ROOT="$( asb::absolute_path "${init_source}" )"
ASB_PROJECT="ansible-service-broker"
SCRIPT_DIR="${ASB_ROOT}/scripts"
TEMPLATE_DIR="${ASB_ROOT}/templates"
BROKER_TEMPLATE="${TEMPLATE_DIR}/deploy-ansible-service-broker-latest.template.yaml"
TEMPLATE_LOCAL_DEV="${TEMPLATE_DIR}/deploy-local-dev-changes.yaml"
GENERATED_BROKER_CONFIG="${ASB_ROOT}/etc/generated_local_development.yaml"

for library_file in $( find "${ASB_ROOT}/scripts/lib" -type f -name '*.sh' -not -path '*/scripts/lib/init.sh' ); do
  source "${library_file}"
done
