# Initial APB Update support

[OBS Update Spec](https://github.com/openservicebrokerapi/servicebroker/blob/v2.12/spec.md#updating-a-service-instance)
[Relevant spec discussion](https://github.com/openservicebrokerapi/servicebroker/issues/139)

## Introduction

Updates allow APB authors to mutate existing service instances.

> By implementing this endpoint, service broker authors can enable users
to modify two attributes of an existing service instance: the service plan
and parameters. By changing the service plan, users can upgrade or downgrade
their service instance to other plans. By modifying properties, users can change
configuration options that are specific to a service or plan

Therefore, the cases we're focused on are:

1) Changing the argument used for a given parameter. Ex: change a "name"

2) Moving from one plan to another.  Ex: Move from non-persistent plan to
a persisted plan. Another example of this could be moving from a multi-tenant
environment to a dedicated environment for a service's database.

NOTE: The Service Catalog currently does **not** implement update, so triggering
these from the UI is not possible. Updates should be considered very "alpha".
To demonstrate , the plan is to enhance `asbcli` to support update.

Updates are a potentially very complex subject considering the full breadth
of what that can mean for any given APB, so the desire here is to control scope
for an initial first-pass of what this could be.

**First-pass constraints**

* Attempting to update a service that has outstanding bindings is invalid and
will result in a 400. Concern here is that a param update or plan transition
could have breaking/destructive consequences for a binding consumer. We think
it's possible in the future to support this, but it deserves an independent
proposal and implementation plan separate from the first-pass.

## Target Demos

### Basic 

* Etherpad - demonstrate an update to initial pad title, foo -> bar
* Etherpad - demonstrate an update from non-persistent plan to persistent plan
* Etherpad - demonstrate an invalid plan update request fails with bad request

### Stretch

* Demo update of credentials, moving a postgresql
database from plan "silver" to "gold" and issuing new admin credentials as part of
the process
* Demo migration of data during a plan transition.
Ex: postgresql "silver" -> "gold" plan, migrating data from "silver" -> "gold"
dbs
* Demo example of a schema migration. Ex: web app moving
from 1.0 -> 2.0 that brings its own DB, and needs to migrate the schema as
part of that version update


## Implementation Details

### Broker

As outlined in the spec, broker must support both async and sync update variants.
Async is requested via `accepts_incomplete` query param.

Broker should prefer to delegate any details of what an "update" actually means
to the APBs. That means passing all the `previous_values` it receives as part
of the update request, along with the set of `next_values` (reading the upstream
spec, the `next_values` are not wrapped in that key specifically). It allows
for the APBs to define "this is how we will move from our current state, to the
next". Example request body taken from upstream spec, let's refer to this as
the `UpdatePayload`:

```json
{
  "context": {
    "platform": "cloudfoundry",
    "some_field": "some-contextual-data"
  },
  "service_id": "service-guid-here",
  "plan_id": "plan-guid-here",
  "parameters": {
    "parameter1": 1,
    "parameter2": "foo"
  },
  "previous_values": {
    "plan_id": "old-plan-guid-here",
    "service_id": "service-guid-here",
    "organization_id": "org-guid-here",
    "space_id": "space-guid-here"
  }
}
```

After validating a given request, the broker will pass this payload through to
an APB, triggering a new "update" action that should execute a new `update.yml`
playbook. See APB section for more.

#### New Credentials

It's possible for param updates or plan transitions to mutate credentials
as a result of an update. Ex: database plan transition needs to issue new
database credentials as a result of the update

NOTE: Since the first-pass is limited to updates of instances without outstanding
bindings, the only possible active credentials in the broker are those from
a provision.

Therefore, the broker *must* monitor an update APB with existing `ext_creds`
functions, looking for newly issued creds.

If credentials are discovered, the newly issued object should be merged with
the existing credentials, and then reinserted into etcd as the new, official
credentials object. Ex:

**original creds**
```json
{
  "db_name": "foobar_db",
  "db_user": "duder",
  "db_pass": "topsecret"
}
```

**updated creds; extracted from update APB run**
```json
{
  "db_user": "duder_prime",
  "db_pass": "supertopsecret"
}
```

**new creds; merge(updated, original)**
```json
{
  "db_name": "foobar_db",
  "db_user": "duder_prime",
  "db_pass": "supertopsecret"
}
```

* Q: Is there any value in versioning this object? Ex: `_apb_creds_version` field
somewhere that begins at 1 and increments each time an update occurs?

#### APB

Overall, the goal with update should be to push the complexity of what an update
actually means for an APB into the APB ansible itself. APBs supporting update
must now define a new playbook similar to the other actions called `update.yml`. This
playbook should define the APB's update behavior given the `UpdatePayload` from
from the broker.

**plan transitions**

The OSB spec explicitly states the broker must reject plan transitions that are
invalid. Effectively, this is a directed graph with transitions as edges and
plans as the nodes. APBs will need to define this graph in their `apb.yml`
so the broker is able to validate a requested transition.

Proposal is to add an optional `updates_to` list on `Plan` objects that indicates
valid transitions. If missing, no transitions are available. Example:

```yaml
name: rhscl-postgresql-apb
# ...snip
plans:
  - name: dev
    description: A single DB server with no storage
    free: true
    metadata: # ...snip
    parameters: *_p
    updates_to:
      - silver
      - gold
  - name: silver
    description: Silver DB plan with persistence
    free: true
    metadata: # ...snip
    parameters: *_p # param anchor
    updates_to:
      - gold
  - name: gold
    description: A single DB server with persistent storage
    free: true
    metadata: # ...snip
    parameters: *_p # param anchor
```

Valid transitions:

```
dev -> silver
dev -> gold
silver -> gold
```

Ultimately, it's up to the APBs to decide whether a given transition is valid
or not. If an APB determines it's okay to allow a destructive downgrade from
"gold" to "dev" and declares it as such, the broker will trust the APB and
will allow that transition.

**parameter updates**
Parameters can be marked as updatable by specifying `updatable: true` on the parameter in the APB Spec.

Parameters will default to `updatable: false` if not specified by the user

An example would look like this:
```yaml
  - name: default
    description: An APB that deploys MediaWiki
    free: True
    metadata:
      displayName: Default
      longDescription: This plan deploys a single mediawiki instance without a DB
      cost: $0.00
    parameters:
      - name: mediawiki_db_schema
        default: mediawiki
        type: string
        title: Mediawiki DB Schema
        required: True
      - name: mediawiki_site_name
        default: MediaWiki
        type: string
        title: Mediawiki Site Name
        required: True
        updatable: True
```

The schema output at /v2/catalog is also updated. Whereas up to this point we have only had parameters returned under create in the schema, parameters that can be updated are now populated under update.

```json
          "free": true,
          "schemas": {
            "service_instance": {
              "create": {
...
             },
              "update": {
                "parameters": {
                  "$schema": "http://json-schema.org/draft-04/schema",
                  "additionalProperties": false,
                  "properties": {
                    "mediawiki_site_name": {
                      "default": "MediaWiki",
                      "title": "Mediawiki Site Name",
                      "type": "string"
                    }
                  },
                  "required": [
                    "mediawiki_site_name"
                  ],
                  "type": "object"
                }
              }
            },
...
```

## Next steps
* Updating instances with outstanding bindings, consider both binding consumers
and producers.
