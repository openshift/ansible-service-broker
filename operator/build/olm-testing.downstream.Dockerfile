FROM alpine as munger

ARG operator_name
ARG broker_name
COPY deploy/olm-catalog/openshift-ansible-service-broker-manifests manifests
RUN sed "s,registry.access.redhat.com/openshift/ose-ansible-service-broker-operator:v4.0.0,$operator_name," -i manifests/4.1/openshiftansibleservicebroker.v4.1.0.csv.yaml
RUN sed "s,registry.access.redhat.com/openshift/ose-ansible-service-broker:v4.0.0,$broker_name," -i manifests/4.1/openshiftansibleservicebroker.v4.1.0.csv.yaml

FROM quay.io/openshift/origin-operator-registry:latest

COPY --from=munger manifests manifests
RUN initializer

ENTRYPOINT ["registry-server"]
CMD ["-t", "/tmp/termination-log.txt"]

