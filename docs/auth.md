### How to configure auth in the broker

The broker now supports authentication.


* how to configure the broker
  * enable basic auth
  * disable basic auth

```yaml
broker:
   ...
   auth:
     - type: basic
       enabled: true
```
* how to configure the broker resource
* how to update the username and password in the secret

    Developer section
* how to add a new auth to the broker
