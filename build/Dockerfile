FROM centos:7
MAINTAINER Ansible Service Broker Community

COPY broker /usr/bin/asbd
COPY entrypoint.sh ansible-service-broker /usr/bin/

RUN useradd -M -s /usr/sbin/nologin -u 1001 -g 0 ansibleapp

RUN mkdir /var/log/ansible-service-broker \
    && touch /var/log/ansible-service-broker/asb.log \
    && mkdir /etc/asb \
    && mkdir -p /usr/local/ansible-service-broker/bin

COPY get_images_for_org.sh /usr/local/ansible-service-broker/bin
COPY config.yaml /etc/ansible-service-broker/config.yaml

RUN chown -R ansibleapp: /var/log/ansible-service-broker \
    && chown -R ansibleapp: /etc/ansible-service-broker

USER ansibleapp

ENTRYPOINT ["entrypoint.sh"]
