FROM quay.io/operator-framework/ansible-operator:master

LABEL com.redhat.delivery.appregistry=true

ADD deploy/olm-catalog/openshift-ansible-service-broker-manifests /manifests

COPY roles/ ${HOME}/roles/
COPY watches.yaml ${HOME}/watches.yaml
COPY playbook.yaml ${HOME}/playbook.yaml
