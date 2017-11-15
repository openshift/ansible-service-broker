# APB Generate Credentials Secret Proposal

## Introduction

This proposal aims to improve the broker process by which we extract
credentials from bindable [Ansible Playbook
Bundles (APB)](https://github.com/ansibleplaybookbundle/ansible-playbook-bundle)
by allowing the APB to generate a secret with needed credentials instead of
using `kubectl exec` into the running pod to grab credentials from a file.

Reference Issue #544

## Problem Description

The broker should seek to use available kubernetes API calls to get necessary
information whenever possible instead of relying on `kubectl exec`. Now that
the broker supplies pod name and namespace information to running APBs via the
[Kubernetes Downward
API](https://kubernetes.io/docs/tasks/inject-data-application/downward-api-volume-expose-pod-information/),
APBs have all they need to generate secrets inside the APB sandbox namespace.
These secrets will be retrieved by the broker after the pod completes
successfully.

## Implementation Details

- [x] Update the [`asb_encode_binding` module](https://github.com/ansibleplaybookbundle/ansible-asb-modules/blob/master/library/asb_encode_binding.py)
  to create a kubernetes secret in the APB sandbox namespace.
  [ansible-asb-modules#8](https://github.com/ansibleplaybookbundle/ansible-asb-modules/pull/8)
- [x] Update
  [`apb-base`](https://github.com/ansibleplaybookbundle/apb-base/tree/master/files/usr/bin):
  we no longer need to run `bind-init` or `broker-bind-creds`.
  [apb-base#7](https://github.com/ansibleplaybookbundle/apb-base/pull/7)
- [x] Add `LABEL "com.redhat.apb.runtime"="2"` to apb-base since this is a
  runtime change.
  ~~Bump the APB version to `2.0`. This will prevent older broker's from grabbing~~
  ~~new APBs that it won't be able to handle and allow us to centrally locate our~~
  ~~backwards compatibility in the broker.~~
  ~~[ansible-playbook-bundle#163](https://github.com/ansibleplaybookbundle/ansible-playbook-bundle/pull/163)~~
- [x] Update broker to evaluate APB runtime versions, support min/max runtime
  versions, and assume runtime version is `1` if label doesn't exist.
- [x] Update
  [`pkg/apb/ext_creds.go
  ExtractCredentials`](https://github.com/openshift/ansible-service-broker/blob/8dda3277/pkg/apb/ext_creds.go#L33)
  **if runtime version `== 1`**: Do it the old way
  **if runtime version `>= 2`**:
  1) watch pod and wait for it to complete
  2) evaluate success/failure of APB execution
  3) read credentials from secret.
- [x] ~~Bump [`MinAPBVersion`] and~~
  ~~[`MaxAPBVersion`](https://github.com/openshift/ansible-service-broker/blob/8dda3277/pkg/version/apbversion.go#L27)~~
  ~~to `2.0`~~
