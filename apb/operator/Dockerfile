FROM quay.io/water-hole/ansible-operator

COPY . /opt/ansible/roles/automation-broker
COPY playbooks/operator.yml /opt/ansible/deploy.yaml
COPY operator/watches.yaml /opt/ansible/watches.yaml
