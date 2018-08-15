FROM ansibleplaybookbundle/apb-base:latest

LABEL "com.redhat.apb.spec"=\
"LS0tCgp2ZXJzaW9uOiAxLjAuMApuYW1lOiBhdXRvbWF0aW9uLWJyb2tlci1hcGIKZGVzY3JpcHRp\
b246IERlcGxveXMgYXV0b21hdGlvbiBicm9rZXIKYmluZGFibGU6ICJGYWxzZSIKYXN5bmM6IG9w\
dGlvbmFsCm1ldGFkYXRhOgogIGRpc3BsYXlOYW1lOiBBdXRvbWF0aW9uIEJyb2tlciAoQVBCKQog\
IGxvbmdEZXNjcmlwdGlvbjoKICAgIEFuIEFQQiBmb3IgbWFuYWdpbmcgdGhlIGF1dG9tYXRpb24g\
YnJva2VyIGluIGEgY2x1c3RlcgogICAgV0FSTklORyAtIFVzZSBvZiB0aGlzIEFQQiB0aHJvdWdo\
IHRoZSBzZXJ2aWNlLWNhdGFsb2cgaXMKICAgIG5vdCBzdXBwb3J0ZWQuCiAgZGVwZW5kZW5jaWVz\
OiBbXQogIHByb3ZpZGVyRGlzcGxheU5hbWU6ICJSZWQgSGF0LCBJbmMuIgpwbGFuczoKICAtIG5h\
bWU6IG1haW4KICAgIGRlc2NyaXB0aW9uOiBNYW5hZ2UgdGhlIGF1dG9tYXRpb24gYnJva2VyIGlu\
IGEgY2x1c3RlcgogICAgZnJlZTogIlRydWUiCiAgICBkZWZhdWx0OiAiVHJ1ZSIKICAgIG1ldGFk\
YXRhOgogICAgICBkaXNwbGF5TmFtZTogRGVmYXVsdAogICAgICBjb3N0OiAkMC4wMAogICAgcGFy\
YW1ldGVyczoKICAgICAgLSBuYW1lOiBicm9rZXJfbmFtZQogICAgICAgIGRlZmF1bHQ6IGF1dG9t\
YXRpb24tYnJva2VyCiAgICAgICAgdHlwZTogc3RyaW5nCiAgICAgICAgZGVzY3JpcHRpb246IE5h\
bWUgb2YgdGhlIGJyb2tlcgogICAgICAgIHRpdGxlOiBCcm9rZXIgTmFtZQogICAgICAgIHJlcXVp\
cmVkOiB0cnVlCiAgICAgIC0gbmFtZTogYnJva2VyX2tpbmQKICAgICAgICBkZWZhdWx0OiBTZXJ2\
aWNlQnJva2VyCiAgICAgICAgdHlwZTogc3RyaW5nCiAgICAgICAgZGVzY3JpcHRpb246IEtpbmQg\
b2YgdGhlIGJyb2tlcgogICAgICAgIHRpdGxlOiBCcm9rZXIgS2luZAogICAgICAgIHJlcXVpcmVk\
OiB0cnVlCg=="

ARG VERSION=latest
ARG APB=latest

RUN yum -y install epel-release openssl && yum clean all

# Add our role into the ansible roles dir
ADD playbooks /opt/apb/actions
ADD . /opt/ansible/roles/automation-broker-apb

# Replace the broker version and apb tag
RUN sed -i "s/\(broker_image_tag:\).*/\1 ${VERSION}/" \
    /opt/ansible/roles/automation-broker-apb/defaults/main.yml
RUN sed -i "s/\(broker_dockerhub_tag:\).*/\1 ${APB}/" \
    /opt/ansible/roles/automation-broker-apb/defaults/main.yml

RUN chmod -R g=u /opt/{ansible,apb}

USER apb
