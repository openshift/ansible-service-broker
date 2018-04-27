FROM centos:7
MAINTAINER Ansible Service Broker Community

ARG DEBUG_PORT=9000

ENV USER_NAME=ansibleservicebroker \
    USER_UID=1001 \
    BASE_DIR=/opt/ansibleservicebroker \
    DEBUG_PORT=${DEBUG_PORT}
ENV HOME=${BASE_DIR}

RUN mkdir -p ${BASE_DIR} ${BASE_DIR}/etc \
 && useradd -u ${USER_UID} -r -g 0 -M -d ${BASE_DIR} -b ${BASE_DIR} -s /sbin/nologin -c "ansibleservicebroker user" ${USER_NAME} \
 && chown -R ${USER_NAME}:0 ${BASE_DIR} \
 && chmod -R g+rw ${BASE_DIR} /etc/passwd


RUN yum -y update \
 && yum -y install epel-release centos-release-openshift-origin \
 && yum -y install origin-clients net-tools bind-utils \
 && yum clean all

RUN mkdir /var/log/ansible-service-broker \
    && touch /var/log/ansible-service-broker/asb.log \
    && mkdir /etc/ansible-service-broker

COPY dev-entrypoint.sh /usr/bin/
COPY dlv /usr/bin/dlv
COPY broker /usr/bin/asbd
COPY migration /usr/bin/migration
COPY dashboard-redirector /usr/bin/dashboard-redirector

RUN chown -R ${USER_NAME}:0 /var/log/ansible-service-broker \
 && chown -R ${USER_NAME}:0 /etc/ansible-service-broker \
 && chmod -R g+rw /var/log/ansible-service-broker /etc/ansible-service-broker

USER ${USER_UID}
RUN sed "s@${USER_NAME}:x:${USER_UID}:@${USER_NAME}:x:\${USER_ID}:@g" /etc/passwd > ${BASE_DIR}/etc/passwd.template

EXPOSE ${DEBUG_PORT}

ENTRYPOINT ["dev-entrypoint.sh"]
