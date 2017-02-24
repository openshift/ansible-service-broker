# asbcli

### Deploying a broker

`asbcli up <cluster-host>:<port> --cluster-user=<user> --dockerhub-user=<user>`

**Example**

`asbcli up cap.example.com:8443 --cluster-user=admin --dockerhub-user=eriknelson`

`asbcli up` will also accept `--cluster-pass` and `--dockerhub-pass`, otherwise,
these will be asked for interactively.

### Connecting to a broker

`asbcli connect <broker-host>:<port>`

**Example**

`asbcli connect asb-1338-ansible-service-broker.cap.example.com`
