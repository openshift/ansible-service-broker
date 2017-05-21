FROM centos:7
MAINTAINER Ansible Service Broker Community

RUN useradd -M -s /usr/sbin/nologin -u 1001 -g 0 ansibleapp

RUN mkdir /var/log/ansible-service-broker \
    && touch /var/log/ansible-service-broker/asb.log \
    && mkdir /etc/ansible-service-broker \
    && mkdir -p /usr/local/ansible-service-broker/bin

RUN chown -R ansibleapp: /var/log/ansible-service-broker \
    && chown -R ansibleapp: /etc/ansible-service-broker

COPY get_images_for_org.sh /usr/local/ansible-service-broker/bin
COPY config.yaml /etc/ansible-service-broker/config.yaml
COPY entrypoint.sh ansible-service-broker /usr/bin/
COPY broker /usr/bin/asbd

USER ansibleapp

ENTRYPOINT ["entrypoint.sh"]
