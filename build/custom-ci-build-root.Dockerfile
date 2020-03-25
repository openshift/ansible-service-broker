FROM openshift/origin-release:golang-1.13

RUN yum install -y epel-release \
    && yum install -y python36-devel python3-pip gcc

RUN pip3 install -U setuptools wheel && pip3 install -U more-itertools==7.0.0 molecule==2.22.0 jmespath openshift

RUN chmod g+rw /etc/passwd
