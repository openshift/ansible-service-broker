FROM openshift/origin-release:golang-1.10

RUN yum install -y epel-release \
    && yum install -y python-devel python-pip gcc

RUN pip install -U setuptools wheel && pip install -U molecule==2.20.0 jmespath openshift



