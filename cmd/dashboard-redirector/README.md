# Dashboard Redirector

**IMPORTANT** This is an alpha feature, meaning it can change at any time!

Currently, the OSB spec does not allow for the brokers to return a `dashboard_url`
at the conclusion of an async provision. This is problematic because brokers
(and bundles, in our case) may not actually know their `dashboard_url` if its
a path that's dynamically determined while the instance is being provisioned.

The issue has been documented in the spec [here](https://github.com/openservicebrokerapi/servicebroker/issues/498).

As a workaround, our broker runs this "dashboard-redirector" as a simple
service that accepts a service instance ID, and looks up the service instance
in our DAO. If the `DashboardURL` is present on the service instance, the
dashboard-redirector will return a 302 redirect to that location. Otherwise,
the redirector will return a 404.

## URL Format

The redirector uses a query parameter for users to request the redirect in the form of:

`<redirector_route>/?id=<service_instance_id>`

Concrete example:

`http://dr-1337-ansible-service-broker.172.17.0.1.nip.io/?id=a5b92aa1-f094-4ffa-a74a-64736f3f48e8`
