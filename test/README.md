# TEST SCRIPTS
Some bash scripts to exercise the OSB API using curl.

* bind.sh - creates a bind
* bootstrap.sh - bootstraps the broker
* catalog.sh - returns the list of APBs
* deprovision.sh - deprovisions an APB
* getbind.sh - returns the bind created by async bind
* getinstance.sh - returns the instance created by provision
* last_operation.sh - polls the last_operation endpoint for job status
* provision-200.sh - sample script to test out the normal error path
* provision-409.sh - sample script to test out the 409 conflict error path
* provision.sh - provisions a service instance
* unbind.sh - deletes a bind
