## Bind / Unbind New Work flow

The goal of this new work flow is to give more control over the actions that the database takes on bind and unbind to the APB author. This means that we want the apb to be launched on on bind and unbind. 

### Default Settings Needed
* set `launch_apb_on_bind: true` - If this is set to false we will behave the exact same as today, handing back the credentials received at provision to the service catalog.


### Brokers Responsibility
* The broker will save the credentials for each bind. this will be saved in etcd under the `extracted_credentials` directory. The key will be the binding id.
* The broker will be responsible for providing the provisioned credentials to the APB as extra vars during bind and unbind. Example
```
--extra-vars '{"provision_params": {}, "bind_params":{}, "db_params:{}'
```
* The broker will be responsible for deleting the bindings extracted credentials in etcd during unbind.
* The broker will be responsible for deleting the bindings extracted credentials in etcd during deprovision 


### APB Responsibility
* The APB is responsible for encoding new or old parameters into `/var/tmp/bind-creds`.
* The APB is responsible for handling the use of superuser vs. regular user. 
* The APB is responsible for handling the deletion/creation of users for their image.