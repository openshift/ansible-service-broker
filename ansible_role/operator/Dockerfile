FROM quay.io/water-hole/ansible-operator
ARG OLM_MANAGED=false

COPY . /opt/ansible/roles/automation-broker
COPY playbooks/operator.yml /opt/ansible/deploy.yml
COPY operator/watches.yaml /opt/ansible/watches.yaml

RUN sed -i "s/\(olm_managed:\).*/\1 ${OLM_MANAGED}/" \
    /opt/ansible/roles/automation-broker/defaults/main.yml
