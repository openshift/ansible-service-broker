FROM centos:7
MAINTAINER Ansible Service Broker Community

RUN curl https://copr.fedorainfracloud.org/coprs/g/ansible-service-broker/ansible-service-broker/repo/epel-7/group_ansible-service-broker-ansible-service-broker-epel-7.repo -o /etc/yum.repos.d/asb.repo

RUN yum -y update \
 && yum -y install epel-release centos-release-openshift-origin \
 && yum -y install origin-clients ansible-service-broker ansible-service-broker-container-scripts \
 && yum clean all

USER ansibleservicebroker

ENTRYPOINT ["entrypoint.sh"]
