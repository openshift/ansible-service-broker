PROJECT="ansible-service-broker"
BROKER_IMAGE="ansibleplaybookbundle/ansible-service-broker:latest"
OPENSHIFT_TARGET="https://kubernetes.default"
OPENSHIFT_USER="admin"
OPENSHIFT_PASS="admin"
DOCKERHUB_USER="CHANGEME"
DOCKERHUB_PASS="CHANGEME"
DOCKERHUB_ORG="ansibleplaybookbundle"
REGISTRY_TYPE="dockerhub"
REGISTRY_URL="docker.io"
DEV_BROKER="true"


VARS="-p BROKER_IMAGE=${BROKER_IMAGE} -p OPENSHIFT_TARGET=${OPENSHIFT_TARGET} -p OPENSHIFT_PASS=${OPENSHIFT_PASS} -p OPENSHIFT_USER=${OPENSHIFT_USER} -p DOCKERHUB_ORG=${DOCKERHUB_ORG} -p DOCKERHUB_PASS=${DOCKERHUB_PASS} -p DOCKERHUB_USER=${DOCKERHUB_USER} -p REGISTRY_TYPE=${REGISTRY_TYPE} -p REGISTRY_URL=${REGISTRY_URL} -p DEV_BROKER=${DEV_BROKER}"

oc delete project ${PROJECT}
oc projects | grep ${PROJECT}
while [ $? -eq 0 ]
do
  echo "Waiting for ${PROJECT} to be deleted"
  sleep 5;
  oc projects | grep ${PROJECT}
done


oc new-project ${PROJECT}
oc process -f deploy-ansible-service-broker.template.yaml -n ${PROJECT} ${VARS}  | oc create -f -
