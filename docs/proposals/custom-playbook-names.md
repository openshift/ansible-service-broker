# Custom Playbook Names

## Introduction
Instead of forcing the user to have a playbook named provision, deprovision,
bind, or unbind, allow the user to create a single playbook and select the
broker action at the role or task level.


## Problem Description
If I want to convert my ansible playbooks to an APB, it's required that I create
provision, bind, unbind, or deprovision playbook(s). This is inconvenient
because I would need to change my playbook name, documentation, possibly some
variables, and break any existing users just so I can support the APB format.

If I want to have multiple APBs in a single git repo I get a similar issue.
In the ```playbooks``` directory, we currently expect a playbook for each
action.  I could change the directory in the Dockerfile to include
```playbooks/apb_number_1```, but it gets back the the initial problem I
outlined where I need to make my playbook structure to fit the APB mold.

Common ansible playbook structure
```bash
playbooks/
├── apb_number_1.yml  # Provision Playbook
├── apb_number_2.yml
└── roles
    ├── apb_number_1
    │   └── provision.yml  # Provision Tasks
    └── apb_number_2
        └── provision.yml  # The ```action.yml``` file is pushed down in the directory structure
```

Not as common ansible playbook structure
```bash
playbooks
├── apb_number_1
│   └── provision.yml   # Provision Playbook
└── apb_number_2
    └── provision.yml
```

Not as common ansible playbook structure
```bash
multiple_apb_repo
├── apb_number_1
│   └── provision.yml   # Provision Playbook
└── apb_number_2
    └── provision.yml
```

The APB structure is rigid and if I want to turn my existing playbook into
an APB I will have to make some changes to the structure of my playbooks.
This will make is difficult for folks who already have playbooks to buy into
the APB's structure.

However, I think we need a good balance of structure and flexibility.  I think
the contract should allow for users to be able to drop in any ansible playbook,
but they will require a few variables, an ```apb.yml```, and a Dockerfile
to have an APB.

## Using ```Name``` in ```apb.yml```
Use ```Name```, in ```apb.yml```, to search for the playbook name to run inside
the APB by passing it in as an evironment variable to the APB as
```BUNDLE_NAME```.

```diff
+BUNDLE_NAME="${BUNDLE_NAME:-}"
...

ANSIBLE_ROLES_PATH=/etc/ansible/roles:/opt/ansible/roles ansible-playbook $playbooks/$ACTION.yaml "${@}" ${extra_args}
elif [[ -e "$playbooks/$ACTION.yml" ]]; then
  ANSIBLE_ROLES_PATH=/etc/ansible/roles:/opt/ansible/roles ansible-playbook $playbooks/$ACTION.yml  "${@}" ${extra_args}
+elif [[ -e "$playbooks/$BUNDLE_NAME.yaml" ]]; then
+  ANSIBLE_ROLES_PATH=/etc/ansible/roles:/opt/ansible/roles ansible-playbook $playbooks/$BUNDLE_NAME.yaml "${@}" ${extra_args} -e action=$ACTION
+elif [[ -e "$playbooks/$BUNDLE_NAME.yml" ]]; then
+  ANSIBLE_ROLES_PATH=/etc/ansible/roles:/opt/ansible/roles ansible-playbook $playbooks/$BUNDLE_NAME.yml "${@}" ${extra_args} -e action=$ACTION
```

## Add a New Field to ```apb.yml```
The new field, ```playbook_name```, would either have to be added to the OSB API
spec or ignored by catalog.

It would get picked up in the broker and passed to the APB as the environment
variable ```BUNDLE_NAME```.

## Work Items
 - Check for ```"$playbooks/$BUNDLE_NAME.yaml"``` in apb-base
 - Add ```BUNDLE_NAME``` to the APB environment
