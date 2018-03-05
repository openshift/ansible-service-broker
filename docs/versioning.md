Versioning Strategy
===================

# Version Format

The Broker version is stored in the broker's RPM spec file,
[ansible-service-broker.spec](../ansible-service-broker.spec), and is in the
form of MAJOR.MINOR.PATCH incremented by the following rules:

* MAJOR version when incompatible API changes are made.
* MINOR version when a new version of
  [openshift/origin](https://github.com/openshift/origin) is being targeted.
* PATCH version when tagged via `tito tag`.

NOTE: The first official release of the broker,
[`1.0.x`](https://github.com/openshift/ansible-service-broker/tree/ansible-service-broker-1.0.1-1),
targeted OpenShift Origin [release
v3.7.0](https://github.com/openshift/origin/releases/tag/v3.7.0). Following the
versioning rules, the version was bumped to `1.1.x` for OpenShift Origin [release
v3.9.0](https://github.com/openshift/origin/tree/release-3.9) and again to
`1.2.x` for OpenShift Origin release v3.10.0.

# Branching and Tagging

All development work is done on the [`master`
branch](https://github.com/openshift/ansible-service-broker/tree/master). New
branches are created when a submitted Pull Request is targeting a future release.
This is done to allow Pull Requests to move forward when new features are no
longer being added to the currently targeted OpenShift releases. When a new
branch is created, it will be named `release-MAJOR.MINOR` and the version in
the `master` branch will be bumped based on the rules in [Version
Format](#VersionFormat).
