# Dockerfiles and Tags
There are three dockerfiles here being used to generate containers for three different tags in the docker.io/ansibleplaybookbundle/origin-ansible-service-broker repo.
- **Canary**: Automated images built from source. These are generally intended to help work on development of the ansible-service-broker and can be expected to break frequently.
- **Latest**: Stable images released less frequently and expected to work with the latest apb containers. These are built using RPM's from the @ansible-service-broker/ansible-service-broker-latest copr repo. The packages in this repo are built using tito when we're fairly confident we'll produce a stable build.
- **Nightly**: Automated image builds using automated RPM builds. This tag is intended to ensure RPM builds work on an ongoing basis. These are built using RPM's from the @ansible-service-broker/ansible-service-broker-nightly copr repo.
- As time goes on and releases are made we may also create tags for specified versions, for example, v1.0, v1.1, etc. In most cases expect that these were retagged from latest.
