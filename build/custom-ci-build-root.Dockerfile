FROM openshift/origin-release:golang-1.13

ENV OPERATOR=/usr/local/bin/ansible-operator \
    USER_UID=1001 \
    USER_NAME=ansible-operator\
    HOME=/opt/ansible

RUN (yum install -y https://dl.fedoraproject.org/pub/epel/epel-release-latest-7.noarch.rpm || true) \
 && yum install -y python-devel gcc inotify-tools \
 && easy_install pip \
 && pip install -U --no-cache-dir setuptools pip \
 && pip install  molecule==2.22 jmespath openshift

# Ensure directory permissions are properly set
RUN echo "${USER_NAME}:x:${USER_UID}:0:${USER_NAME} user:${HOME}:/sbin/nologin" >> /etc/passwd
RUN mkdir -p ${HOME}/.ansible/tmp \
 && chown -R ${USER_UID}:0 ${HOME} \
 && chmod -R ug+rwx ${HOME}
