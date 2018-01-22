# TEST SCRIPTS
Some bash scripts to exercise the OSB API using curl.

* bind.sh - creates a bind
* bootstrap.sh - bootstraps the broker
* catalog.sh - returns the list of APBs
* deprovision.sh - deprovisions an APB
* getbind.sh - returns the bind created by async bind
* getinstance.sh - returns the instance created by provision
* last_operation.sh - polls the last_operation endpoint for job status
* provision.sh - provisions a service instance
* unbind.sh - deletes a bind

When running the scripts you will need to be logged into the OpenShift cluster
since the scripts will use `oc whoami -t` to get your token for authentication.

NOTE: and you can't be a system:admin for this.

NOTE: You will also need to enable `auto_escalate` feature on the broker. You
can do this by editing the template or by editing the configmap of an already
deployed broker
