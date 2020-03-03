FROM openshift/origin-release:golang-1.13

ENV OPERATOR=/usr/local/bin/ansible-operator \
    USER_UID=1001 \
    USER_NAME=ansible-operator\
    HOME=/opt/ansible

# Install python dependencies
# Ensure fresh metadata rather than cached metadata in the base by running
# yum clean all && rm -rf /var/yum/cache/* first
RUN yum clean all && rm -rf /var/cache/yum/* \
 && yum -y update \
 && yum install -y libffi-devel openssl-devel python36-devel gcc python3-pip python3-setuptools \
 && FEDORA=$(case $(arch) in ppc64le|s390x) echo -n fedora-secondary ;; *) echo -n fedora/linux ;; esac) \
 && yum install -y https://dl.fedoraproject.org/pub/$FEDORA/releases/30/Everything/$(arch)/os/Packages/i/inotify-tools-3.14-16.fc30.$(arch).rpm \
 && pip3 install --no-cache-dir --ignore-installed ipaddress \
      ansible-runner==1.3.4 \
      ansible-runner-http==1.0.0 \
      openshift==0.9.2 \
      ansible~=2.9 \
      jmespath \
 && yum remove -y gcc libffi-devel openssl-devel python36-devel \
 && yum clean all \
 && rm -rf /var/cache/yum \
 && ansible-galaxy collection install operator_sdk.util

# Ensure directory permissions are properly set
RUN echo "${USER_NAME}:x:${USER_UID}:0:${USER_NAME} user:${HOME}:/sbin/nologin" >> /etc/passwd
RUN mkdir -p ${HOME}/.ansible/tmp \
 && chown -R ${USER_UID}:0 ${HOME} \
 && chmod -R ug+rwx ${HOME}
