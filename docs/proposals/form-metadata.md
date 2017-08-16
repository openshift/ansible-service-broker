# Add UI metadata to catalog requests

## Introduction
Support better provision user experience for APBs by giving additional metadata about APB parameters.

## Problem Description
Currently, APB parameters are shown in random order and without any kind of grouping.  the OpenServiceBroker API utilizes json-schema data format and does not include any UI metadata.  As such, the UI cannot customize the experience and users are confronted with a bad experience which gets worse as the number of parameters increases. 

## Implementation Details

### APB changes
APBs will allow an author to implement parameter order, grouping, and form field type.
* Ordering will honor the order of parameters used in the apb.yml
* Grouping will require the addition of a parameter field `display_group`.  This is an arbitrary string value but can be left empty to signify that the form input is at the top level and not grouped.  If it matches other `display_group` string from a parameter immediately before or after the parameter in question, they would be grouped together.
* UI field types will require the addition of a parameter field `display_type`.  Default (empty) will use the default form input based on the json-schema.

```yaml
name: hello-world-db-apb
# ...snip...
plans:
  - name: default
    # ...snip...
    parameters:                           # respects order of parameters
      - name: postgresql_database         # first parameter on form
        title: PostgreSQL Database Name
        type: string
        required: True
        default: admin
      - name: postgresql_user             # second parameter
        title: PostgreSQL User
        type: string
        required: True
        default: admin
        display_group: User Information   # same group as password       
      - name: postgresql_password         # third parameter
        title: PostgreSQL Password
        type: string
        required: True
        default: admin
        display_type: password            # password type displayed in UI as ****
        display_group: User Information   # same group as user
```
Documentation for the additional fields should be added to the appropriate documentation for APBs in the [ansible-playbook-bundle](https://github.com/fusor/ansible-playbook-bundle) project.

### API changes
The agreed upon request format with the OpenShift UI team matches the [form definition](https://github.com/json-schema-form/angular-schema-form/blob/development/docs/index.md#form-definitions) used on the front end UI.  This will be specific to OpenShift but there is no standard agreed upon and this would cause the least changes in the UI code with maximum flexibility.  In order to minimize impact and not make changes to the OpenServiceBroker API, the data will be contained in the plan metadata.  Since parameters are tied to the method, the json structure in metadata containing the form definition will mirror the json structure containing the schema.

```json
[
  {
    "name": "hello-world-db-apb-latest",
    "metadata": {},
    "plans": [
      {
        "id": "default",
        "name": "default",
        "metadata": {
          "schemas": {
            "service_instance": {
              "create": {
                "form_definition": [
                  "postgresql_database",
                  {
                    "type": "section",
                    "items": [
                      "postgresql_user",
                      {
                        "key": "postgresql_password",
                        "type": "password"
                      }
                    ]
                  }
                ]
              },
              "update": {}
            },
            "service_binding": {
              "create": "... same as service_instance.create"
            }
          }
        },
        "schemas": {
          "service_instance": {
            "create": {
              "parameters": {
                "$schema": "http://json-schema.org/draft-04/schema",
                "additionalProperties": false,
                "properties": {
                  "postgresql_database": {
                    "default": "admin",
                    "title": "PostgreSQL Database Name",
                    "type": "string"
                  },
                  "postgresql_password": {
                    "default": "admin",
                    "title": "PostgreSQL Password",
                    "type": "string"
                  },
                  "postgresql_user": {
                    "default": "admin",
                    "title": "PostgreSQL User",
                    "type": "string"
                  }
                },
                "required": [
                  "postgresql_database",
                  "postgresql_user",
                  "postgresql_password"
                ],
                "type": "object"
              }
            },
            "update": {}
          },
          "service_binding": {
            "create": "... same as service_instance.create"
          }
        }
      }
    ]
  }
]

```
Currently the impact would be change the metadata and is additive with no breaking changes.  The broker code would modify how the metadata is set [here](https://github.com/openshift/ansible-service-broker/blob/master/pkg/broker/util.go#L36) and introducing a new function to append the UI metadata.

## Notes
* This will require changes to [origin-web-catalog](https://github.com/openshift/origin-web-catalog) to read, validate, and use the form definition.  Currently we are only planning on supporting only fieldset/sections for grouping, types, and honoring the order from the form definition.

