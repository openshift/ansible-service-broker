# Passing parameters through Secrets

Parameters can be passed through secrets using rules specified in the [broker's config](config.md).

These rules can be added to the broker config automatically using the script scripts/create_broker_secret.py.
This script will also create the desired secret in the broker namespace, and rollout a new broker if the config
has changed.


## Example

Running:
```bash
./scripts/create_broker_secret.py test ansible-service-broker docker.io/ansibleplaybookbundle/rhscl-postgresql-apb  postgresql_user=test postgresql_password=testpassword postgresql_version="9.5"
```
Will create the following secret:

```yaml
---
apiVersion: v1
kind: Secret
metadata:
    name: test
    namespace: ansible-service-broker
stringData:
    "postgresql_version": "9.5"
    "postgresql_user": "test"
    "postgresql_password": "testpassword"
```

and add the following section to the broker config:

```yaml
secrets:
- apb_name: dh-ansibleplaybookbundle-rhscl-postgresql-apb-latest
  secret: test
  title: test
```

When the `docker.io/ansibleplaybookbundle/rhscl-postgresql-apb:latest` APB is run from the catalog UI,
the `postgresql_user`, `postgresql_version`, and `postgresql_password` variables will not be displayed
to the user. Instead, when the APB is run, it will run in the namespace of the broker, and the `test`
Secret will be mounted, parsed, and passed through to the deployment playbook.


## Example using a file

The secret can also be created using a file, instead of passing the parameters through
the command line.
To create the secret using a file, just create a file with the following structure:

parameters.yml
```yaml
---

postgresql_version: 9.5
postgresql_user: test
postgresql_password: testpassword
```

Then run:
```bash
./scripts/create_broker_secret.py test ansible-service-broker docker.io/ansibleplaybookbundle/rhscl-postgresql-apb @parameters.yml
```

This will create the exact same secret and update to the broker configuration as the `key=value` method.


## Note
If a secret is created and added to the broker, the catalog UI will not update to reflect it until it refreshes its list of ServiceClasses.
If the UI is not updated it is likely that the catalog has not made a new `/catalog` request yet.
