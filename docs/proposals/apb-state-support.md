# Abstract

This proposal will outline a simple and secure method for allowing APB developers to set state during a APB
action and then be able to access that state during a subsequent APB action.

## Actions
   - provision
   - deprovision
   - update
   - bind 
   - unbind
   
   
## Problem Description

APBs are stateless. All state is managed for them by the broker. Currently the broker passes in parameters specified
during a request to the catalog. It also passes in some additional parameters such as the namespace etc, additionally 
we also pass in credentials created during the provision. While this is useful, there is no mechanism for an APB to store 
and access data across actions without exposing it to the end user. We want to avoid APB developers working around this 
limitation by doing suboptimal things such as storing extra data in the credentials or by creating a ConfigMap in the users 
namespace as both of these expose the data to the user and as the namespace is controlled by the user, it is naturally 
not trustworthy or reliable.



# Proposed Solution

## APB module for handling state

This module will expose an API to the APB developer for setting and getting key value pairs.

```

- name: Save some stuff
  asb_set_state:
    service_name: "{{ service_name }}"

```

Under the hood this APB module would take the key value pair, and store it in a
ConfigMap (or possibly a secret) labelled ```apb:state```. This ConfigMap would live within the 
temporary namespace where the APB was running. 

## Update broker to manage APB created state ConfigMaps

To ensure the state was persisted across APB action, the broker would watch for ConfigMaps being added in the 
APB namespace with the expected label. When a new ConfigMap was created in the APB's namespace, the broker would 
copy this ConfigMap over to the broker's namepsace and name it ```<ServiceInstanceID>-state```. If a ConfigMap with 
that name was already present, the broker would update and append the values. The broker would also watch 
for ``Modified`` events on the ConfigMap and ensure to update the corresponding ConfigMap in the broker's namespace. 
There should only ever be one ConfigMap per ServiceInstance. The ConfigMap would be removed from the broker's namespace 
once the ServiceInstance was deleted. 

## Update broker to pass through initial state to APB

For every APB action, except provision as there would be no state at this point, if a ConfigMap (ServiceInstanceID-state) is present, 
in the broker's namespace, its key value pairs will be passed through to the APB prefixed with ```state_<key> = value ```
  
    
   