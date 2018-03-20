# Abstract

This proposal will outline a simple and secure method for allowing Service Bundle developers to set state during a Service Bundle
action and then be able to access that state during a subsequent Service Bundle action.

## Actions
   - provision
   - deprovision
   - update
   - bind
   - unbind


## Problem Description

Service Bundles are stateless. All state is managed for them by the broker. Currently, the broker passes in parameters specified
during a request to the catalog. It also passes in some additional parameters such as the namespace etc, additionally,
we also pass in credentials created during the provision. While this is useful, there is no mechanism for a Service Bundle to store
and access data across actions without exposing it to the end user. We want to avoid Service Bundle developers working around this
limitation by doing suboptimal things such as storing extra data in the credentials or by creating a ConfigMap in the user's
namespace as both of these expose the data to the user and, as the namespace is controlled by the user, it is naturally
not trustworthy or reliable.



# Proposed Solution

## Service Bundle contract for handling state

### Add APB specific module

This module will expose an API to the APB developer for setting key-value pairs.

```
- name: save some state
  asb_set_state:
    key: "{{ value }}"


- name: get some state
  asb_get_state:
    key: "key"

```

Under the hood, this APB module would take the key-value pair, and store it in a
ConfigMap named ```$POD_NAME```. This ConfigMap would live within the
temporary namespace ```$POD_NAMESPACE``` and be created by the Service Bundle when with the first call to ```asb_set_state```.

## Update broker to manage Service Bundle created state ConfigMaps

After an action was successfully completed (ie the Service Bundle exited with a 0 exit code) and before the sandbox namespace was removed, the broker would copy the ConfigMap back to the broker's namespace and name it ```<ServiceInstanceID>-state```. If a ConfigMap with
that name was already present, the broker would update and append the values.
There should only ever be one ConfigMap per ServiceInstance. The ConfigMap would be removed from the broker's namespace
once the the deprovision action was invoked for that ServiceInstance.

## Update broker to pass through initial state to Service Bundle

For every Service Bundle action, except provision as there would be no state at this point, if a ConfigMap (ServiceInstanceID-state) is present,
in the broker's namespace, it will be mounted in the ServiceBundle's pod ready to be accessed via the proposed ```apb_get_state``` or by some other manner beyond the scope of this proposal. 
