FROM quay.io/water-hole/ansible-operator

ENV OPERATOR=/usr/local/bin/ansible-operator \
    USER_UID=1001 \
    USER_NAME=operator \
    BASE_DIR=/opt/ansible
ENV HOME=${BASE_DIR}


COPY . /opt/ansible/roles/automation-broker
COPY playbooks/operator.yml /opt/ansible/deploy.yml
COPY operator/config.yaml /opt/ansible/config.yaml

COPY operator/entrypoint.sh /usr/local/bin/entrypoint.sh

RUN mkdir -p /opt/ansible ${BASE_DIR}/.ansible/tmp \
 # && useradd -u ${USER_UID} -r -g 0 -M -d ${BASE_DIR} -b ${BASE_DIR} -s /sbin/nologin -c "operator user" ${USER_NAME} \
 && chown -R ${USER_NAME}:0 ${BASE_DIR} \
 && chmod -R g=u  ${BASE_DIR} /etc/passwd

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
USER ${USER_NAME}
