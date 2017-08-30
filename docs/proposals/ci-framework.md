# Continuous Integration Framework

## Introduction
The current CI infrastructure uses the services: MediaWiki and Postgresql to
test provision, bind, unbind, and deprovision on every PR to the
ansible-service-broker repo. This infrastructure needs to scale so that it can
test provision, bind, unbind, and deprovision on any set of APB(s).

## Problem Description
The four actions that make up the APB API are constant across all APBs. In
order to scale the CI infrastructure, there needs to be a tool that gathers
user input and translates that into one or more APB actions instead of a bash
script that only works for MediaWiki and Postgresql.

## Framework Overview
The new framework could be written in: bash, go, or python. Although bash worked
great for the initial gate work, go or python will be nicer for building an API.

The goal is *not* to build something complex, but to provide an out of the box
framework that will allow an APB developer to easily standup CI in their APB git
repos using Travis or Jenkins.

### Input
Each APB repo will have a config file for running a CI job.

Example input file for the mediawkik apb.
```yaml
 - provision: ansibleplaybookbundle/mediawiki123-apb
 - provision: ansibleplaybookbundle/postgresql-apb
 # Verify provision was successful

 - bind: ansibleplaybookbundle/postgresql-apb
 # Verify bind was successful

 - verify: verify-app.sh
 # Run as a shell command

 - unbind: ansibleplaybookbundle/postgresql-apb
 # Automatic verification occurs after the unbind

 - deprovision: ansibleplaybookbundle/mediawiki123-apb
 - deprovision: ansibleplaybookbundle/postgresql-apb
 # Automatic verification occurs after the deprovision
```

### Output
The execution of CI depends on the input order.

## Code
The functions from ```scripts/broker-ci/local-ci.sh``` will be translated into
API objects with the config file as inputs.

```bash
function provision {...}

function deprovision {...}

function bind {...}

function unbind {...}
```

The Verify function will execute the input command in a shell.  It will expect
the scripts will return 0 for success and 1 for error.

```go
func verify(input ...) {
     output, err := os.Run(input)
}
```

## APB Health
With a new CI framework, there can be a requirement that every new APB have a CI
job in place before being considered 'stable'.  Also, the health of the APB can
be tracked by whether the CI is green or not.

## Infrastructure
This framework is not going to create infrastructure.  There are already tools
that do that.  But, there will be a script or tool available that will setup
infrastructure for testing.  We can start by using
```scripts.broker-ci/setup.sh``` and improve it later.

## Local Testing
```make ci``` will keep it's functionality, but under the covers it will
consume this new framework.

## Work Items
- Build the CI framework.
- Migrate broker ci and ```make ci``` over to the new framework.
- Write docs describing how to use the framework.
