FROM centos:7

ENV HOME=/home/molecule \
    USER_UID=1001
ENV PYTHONUSERBASE=${HOME}/.local \
    PATH="${HOME}/.local/bin:${PATH}"

RUN mkdir -p ${HOME}

WORKDIR ${HOME}

RUN yum install -y epel-release \
    && yum install -y python-devel python-pip gcc

RUN pip install --user -U setuptools && pip install --user molecule==2.20.0.0a2 jmespath openshift

COPY . ${HOME}

RUN chown ${USER_UID}:0 ${HOME} \
 && chmod -R ug+rwx ${HOME} \
 && chmod g+rw /etc/passwd

COPY build/entrypoint.sh /usr/local/bin/entrypoint

ENTRYPOINT ["/usr/local/bin/entrypoint"]
