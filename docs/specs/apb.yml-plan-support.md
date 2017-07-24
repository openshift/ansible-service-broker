# Multi-plan support

## Broker changes

* Broker previously hid the concept of plans from APBs by faking it and generating
a "Default" plan . It entirely ignored the `PlanID` on requests.
Now the Broker should read the `PlanID` on the request and inject it into the parameters
along with the rest of the user provided params.
* Use this as an opportunity to establish a namespace on parameters that the
broker can use for standard, special parameters it sends into APBs. `_apb_<param>`.
In this case, PlanID is passed to APBs as `_apb_plan_id`. APBs should be written
expecting this parameter and branch behavior based on its value.
* Enforce validations at the registry level. If a validation fails, warn spec
is not conformant and filter from available set. Validations:
```
-> len(spec.Plans) > 0
-> plan.name must be unique amongst all plans
```

## APB changes

1) New `plans` section of type `[]Plan`. **At least one plan is **required**
2) Move `parameters` from top level onto `Plan`s, type. New types:

**Plan**

Field Name | Type | Required | Default | Description
---|---|---|---|---
name | string| yes |  | Name of the plan, used as an identifier. **Must be unique amongt plans on the apb**
description | string | yes | | A human readable description of the plan.
free | bool | yes | | Indicates whether the plan has an associated cost
metadata | `PlanMetadata` |  no | | Plan metadata
parameters | `[]Parameter` |  no | | Plan parameters

**PlanMetadata**

Field Name | Type | Required | Default | Description
---|---|---|---|---
displayName | string | no | "" | Name of plan for display purposes
longDescription | string | no | "" | A detailed description of the plan
cost | string | no | "" | Cost string used for display purposes only

---

3) Move any example of a `[]map[string]T` to `[]T`. The former introduced an entirely
unnecessary map layer. They started to get nested after moving params onto plans and
things got very ugly because of it. Reintroduce `name` onto `T`s.

**Old Format**

```yaml
parameters:
  - mediawiki_db_schema:
      default: mediawiki
      type: string
      title: Mediawiki DB Schema

  - mediawiki_site_name:
      default: MediaWiki
      type: string
      title: Mediawiki Site Name
```

**New Format**

```yaml
parameters:
  - name: mediawiki_db_schema
    default: mediawiki
    type: string
    title: Mediawiki DB Schema
  - name: mediawiki_site_name
    default: MediaWiki
    type: string
    title: Mediawiki Site Name
```

Should use this pattern for any additional changes moving forward.

---

4) Remove `required` section. `required` field should return to being an attribute on `Parameter`s.

```yaml
parameters:
  - name: mediawiki_db_schema
    required: true
    default: mediawiki
    type: string
    title: Mediawiki DB Schema
```

## Notes

* Use yaml anchors and inheritence to cut down on parameter duplication if plans share
parameters, see full example.

## Questions

* Obviously this is a large, breaking change to **all** existing APBs and their `apb.yml` files.
How do we want to make across the board so things get rolled out as smoothly as possible?
* Is this a good opportunity to have the broker start respecting the version label found on
apbs? I think it makes sense to have that in lock setup with broker release versions. Is
broker master considered `0.10` since it's `release-0.9++`?

## Full examples

### Single plan example

```yaml
name: rhscl-postgresql-apb
image: ansibleplaybookbundle/rhscl-postgresql-apb
description: SCL PostgreSQL apb implementation
bindable: true
async: optional
metadata:
  documentationUrl: https://www.postgresql.org/docs/
  longDescription: An apb that deploys postgresql 9.4 or 9.5.
  dependencies: ['registry.access.redhat.com/rhscl/postgresql-95-rhel7']
  displayName: PostgreSQL (APB)
  console.openshift.io/iconClass: icon-postgresql
dependencies:
  - postgresql_version
plans:
  - name: default
    description: Postgresql DB
    free: true
    metadata:
      displayName: Default Postgresql DB
      longDescription: This plan provides a simple PostgreSQL server with persistent storage
      cost: $0.00
    parameters:
      - name: postgresql_database
        default: admin
        type: string
        title: PostgreSQL Database Name
        required: true
      - name: postgresql_password
        type: string
        description: A random alphanumeric string if left blank
        title: PostgreSQL Password
      - name: postgresql_user
        default: admin
        title: PostgreSQL User
        type: string
        maxlength: 63
        required: true
      - name: postgresql_version
        default: 9.5
        enum: ['9.5', '9.4']
        type: enum
        title: PostgreSQL Version
        required: true
```

### Shared param set

```yaml
################################################################################
# Shared Parameters
################################################################################
_p: &_p
  - name: postgresql_database
    default: admin
    type: string
    title: PostgreSQL Database Name
    required: true
  - name: postgresql_password
    type: string
    description: A random alphanumeric string if left blank
    title: PostgreSQL Password
  - name: postgresql_user
    default: admin
    title: PostgreSQL User
    type: string
    maxlength: 63
    required: true
  - name: postgresql_version
    default: 9.5
    enum: ['9.5', '9.4']
    type: enum
    title: PostgreSQL Version
    required: true
################################################################################
name: rhscl-postgresql-apb
image: eriknelson/rhscl-postgresql-apb
description: SCL PostgreSQL apb implementation
bindable: True
async: optional
metadata:
  documentationUrl: https://www.postgresql.org/docs/
  longDescription: An apb that deploys postgresql 9.4 or 9.5.
  dependencies: ['registry.access.redhat.com/rhscl/postgresql-95-rhel7']
  displayName: PostgreSQL (APB)
  console.openshift.io/iconClass: icon-postgresql
dependencies:
  - postgresql_version
plans:
  - name: dev
    description: A single DB server with no storage
    free: true
    metadata:
      displayName: Development
      longDescription: This plan provides a single non-HA PostgreSQL server without persistent storage
      cost: $0.00
    parameters: *_p
  - name: prod
    description: HA DB Server with 1TB of Storage
    free: false
    metadata:
      displayName: Production
      longDescription: This plan provides a single non-HA PostgreSQL server with persistent storage
      cost: $5.99 monthly
    parameters: *_p
```
