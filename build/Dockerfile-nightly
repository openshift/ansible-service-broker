FROM centos:7
MAINTAINER Ansible Service Broker Community

RUN curl https://copr.fedorainfracloud.org/coprs/g/ansible-service-broker/ansible-service-broker-nightly/repo/epel-7/group_ansible-service-broker-ansible-service-broker-nightly-epel-7.repo -o /etc/yum.repos.d/asb.repo

ENV USER_NAME=ansibleservicebroker \
    USER_UID=1001 \
    BASE_DIR=/var/lib/ansibleservicebroker
ENV HOME=${BASE_DIR}

RUN yum -y update \
 && yum -y install epel-release centos-release-openshift-origin \
 && yum -y install origin-clients ansible-service-broker ansible-service-broker-container-scripts \
 && yum clean all

RUN mkdir -p ${BASE_DIR} ${BASE_DIR}/etc /var/run/asb-auth \
 && userdel ansibleservicebroker \
 && useradd -u ${USER_UID} -r -g 0 -M -d ${BASE_DIR} -b ${BASE_DIR} -s /sbin/nologin -c "ansibleservicebroker user" ${USER_NAME} \
 && chown -R ${USER_NAME}:0 ${BASE_DIR} /var/log/ansible-service-broker /etc/ansible-service-broker /var/run/asb-auth \
 && chmod -R g+rw ${BASE_DIR} /etc/passwd /etc/ansible-service-broker /var/log/ansible-service-broker /var/run/asb-auth


USER ${USER_UID}
RUN sed "s@${USER_NAME}:x:${USER_UID}:@${USER_NAME}:x:\${USER_ID}:@g" /etc/passwd > ${BASE_DIR}/etc/passwd.template

ENTRYPOINT ["entrypoint.sh"]
