FROM openshift/origin-release:golang-1.13

RUN yum install -y epel-release \
    && yum install -y python-devel python-pip gcc

RUN pip install -U setuptools wheel more-itertools==5.0.0 && pip install  molecule==2.20.0 jmespath openshift

RUN chmod g+rw /etc/passwd

