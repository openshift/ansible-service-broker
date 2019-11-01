FROM quay.io/operator-framework/ansible-operator:master

USER 0
RUN if [[ -f "/usr/local/bin/user_setup" ]] ; then /usr/local/bin/user_setup ; fi
USER 1001

COPY roles/ ${HOME}/roles/
COPY watches.yaml ${HOME}/watches.yaml
COPY playbook.yaml ${HOME}/playbook.yaml
