FROM ansibleplaybookbundle/apb-base:latest

ARG VERSION=latest
ARG APB=latest

RUN yum -y install epel-release openssl && yum clean all

# Add our role into the ansible roles dir
ADD playbooks /opt/apb/actions
ADD . /opt/ansible/roles/automation-broker-apb

# Replace the broker version and apb tag
RUN sed -i "s/\(broker_image_tag:\).*/\1 ${VERSION}/" \
    /opt/ansible/roles/automation-broker-apb/defaults/main.yml
RUN sed -i "s/\(broker_dockerhub_tag:\).*/\1 ${APB}/" \
    /opt/ansible/roles/automation-broker-apb/defaults/main.yml

RUN chmod -R g=u /opt/{ansible,apb}

USER apb
