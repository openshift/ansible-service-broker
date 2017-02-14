##  external storage
* external gluster system
* provision ansibleapp that requires access to gluster resource

## database that handles binds
* provision database ansibleapp (stays up to let other bind)
* bind request to app returns connection information,

    ```
    User -> ServiceCatalog: POST instance
    ServiceCatalog -> Ansible Service Broker: PUT provision/instance_id
    Ansible Service Broker -> etcd : get database image
    etcd -> Ansible Service Broker: return image record
    Ansible Service Broker -> Docker Hub: pull database image
    Docker Hub -> Ansible Service Broker: return database image
    Ansible Service Broker -> Ansible Service Broker: run database image
    Ansible Service Broker -> ServiceCatalog: return 200 image
    ServiceCatalog -> User: ServiceClass
    User -> ServiceCatalog: POST binding
    ServiceCatalog -> Ansible Service Broker: PUT bind
    Ansible Service Broker -> ServiceCatalog: return database connection string
    ServiceCatalog -> ServiceCatalog: Create Binding
    ServiceCatalog -> User: binding instance
    ```
## Etherpad wants to connect to database
* provision etherpad
* bound to database

    sounds like the database if it exists in the same namespace will be INJECTED
    into the etherpad provision as env variables
    ```
    # assume database instance was previously provisioned
    User -> ServiceCatalog: POST etherpad instance
    ServiceCatalog -> Ansible Service Broker: PUT provision/instance_id
    Ansible Service Broker -> etcd : get etherpad image
    etcd -> Ansible Service Broker: return image record
    Ansible Service Broker -> Docker Hub: pull etherpad image
    Docker Hub -> Ansible Service Broker: return etherpad image
    Ansible Service Broker -> Ansible Service Broker: run etherpad image
    Ansible Service Broker -> ServiceCatalog: return 200 image
    ServiceCatalog -> User: ServiceClass
    User -> ServiceCatalog: POST binding
    ServiceCatalog -> Ansible Service Broker: PUT bind
    Ansible Service Broker -> ServiceCatalog: return database connection string
    ServiceCatalog -> ServiceCatalog: Create Binding
    ServiceCatalog -> User: binding instance
    ```
