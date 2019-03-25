FROM openshift/origin-release:golang-1.10

RUN yum install -y epel-release \
    && yum install -y python-devel python-pip gcc

RUN pip install -U setuptools && pip install molecule==2.20.0.0a2 jmespath openshift


RUN chmod g+rw /etc/passwd

