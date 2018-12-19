FROM quay.io/water-hole/ansible-operator
ARG OLM_MANAGED=true
ARG VERSION=v3.11
ARG APB=v3.11

COPY . /opt/ansible/roles/automation-broker
COPY playbooks/operator.yml /opt/ansible/deploy.yml
COPY operator/watches.yaml /opt/ansible/watches.yaml

USER root
RUN sed -i "s/\(broker_image_tag:\).*/\1 ${VERSION}/" \
    /opt/ansible/roles/automation-broker/defaults/main.yml && \
    sed -i "s/\(broker_dockerhub_tag:\).*/\1 ${APB}/" \
    /opt/ansible/roles/automation-broker/defaults/main.yml && \
    sed -i "s/\(olm_managed:\).*/\1 ${OLM_MANAGED}/" \
    /opt/ansible/roles/automation-broker/defaults/main.yml
USER ansible-operator
