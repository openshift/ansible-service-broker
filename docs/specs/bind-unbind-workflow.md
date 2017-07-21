## Bind / Unbind New Work flow

The goal of this new work flow is to give more control over the actions that the database takes on bind and unbind to the APB author. This means that we want the apb to be launched and to run the bind or unbind actions. 

### Default Settings Needed
* set `launch_apb_on_bind: true` - If this is set to false we will behave the exact same as today, handing back the credentials received at provision to the service catalog.


### Brokers Responsibility
* The broker will save the credentials for each bind. this will be saved in etcd under the `extracted_credentials` directory. The key will be the binding id.
* The broker will be responsible for providing the provisioned credentials (extracted credentials given at provision time) to the APB as extra vars during bind and unbind. Example
* `--extra-vars '{"provision_params": {}, "bind_params":{}, "db_params:{}`
  - Note: provision_params and bind_params are currently being defined in the APB and the service catalog is passing them to the broker, to be sent along to the APB.
  - Example: on APB provision, the APB decides to create an admin user & password. This admin user should not be given out during binding. Instead the APB author would like to create new user accounts (by using the already saved/created admin user & password), and pass those back as credentials during the bind. The same is true for unbind. The APB author could want to delete the bind credentials and would need to use the admin user & password. 
  - Example: APB author would like to keep only 1 account, and use that 1 account for all the bindings(current behavior). APB authors can make no guarantees on how the broker is running. (launch_apb_on_bind is true or false). That means their bind would need to pass back the credentials created at provision time. They can now do this by just taking the parameters passed in and sending them back to the broker. 
* The broker will be responsible for deleting the bindings extracted credentials in etcd during unbind.
* The broker will be responsible for deleting the bindings extracted credentials in etcd during deprovision.


### APB Responsibility
* The APB is responsible for encoding new or already created parameters into `/var/tmp/bind-creds`.
* The APB is responsible for handling the use of superuser/admin user vs. regular user to manage their image. 
* The APB is responsible for handling the deletion/creation of users for their image.
