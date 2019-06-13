FROM openshift/origin-release:golang-1.11 AS builder
COPY . /go/src/github.com/openshift/ansible-service-broker
RUN cd /go/src/github.com/openshift/ansible-service-broker \
  && make broker \
  && make dashboard-redirector

FROM registry.svc.ci.openshift.org/openshift/origin-v4.0:base

COPY --from=builder /go/src/github.com/openshift/ansible-service-broker/broker /usr/local/bin/asbd
COPY --from=builder /go/src/github.com/openshift/ansible-service-broker/dashboard-redirector /usr/local/bin/dashboard-redirector
COPY --from=builder /go/src/github.com/openshift/ansible-service-broker/build/entrypoint.sh /usr/local/bin/entrypoint

ENTRYPOINT ["/usr/local/bin/entrypoint"]
USER ${USER_UID}

