# Bind Parameters

## Introduction
When creating a binding from the OpenShift UI, users are presented with available parameters for edit.  Ansible Service Broker (ASB) does not currently support these fields.

## Problem Description
Ansible Service Broker (ASB) must allow users to construct Ansible Playbook Bundles (APBs) that use the binding parameter fields in the UI and supply those fields to an APB when executing the bind action

## APB changes
APBs will have another section named `bind_parameters`
```yaml
name: hello-world-db-apb
description: A sample APB which deploys Hello World Database
bindable: True
async: optional
metadata:
  displayName: Hello World Database (APB)
plans:
  - name: default
    description: A sample APB which deploys Hello World Database
    parameters:
      - name: postgresql_admin_password
        title: PostgreSQL Admin Password
        type: string
        default: admin
        required: true
    bind_parameters:
      - name: postgresql_database
        title: PostgreSQL Database Name
        type: string
      - name: postgresql_user
        title: PostgreSQL User
        type: string
      - name: postgresql_password
        title: PostgreSQL Password
        type: string
```

## ASB Changes
ASB will parse the `bind_parameters` and add them to the plan.  The catalog request should use the `bind_parameters` and send them as part of the service instances, allowing the OpenShift UI to use them.  The form definitions must also be supplied for the UI.  When a binding is created, ASB should capture the parameters from the request and send them to the bind APB as `extra-vars` available for use.

## Work Items
* ASB Changes
* ASB Unit tests
* OpenShift UI form definition changes
* Experimental APB with async bind and passed parameters
* APB documentation changes (getting started / developer guide)

