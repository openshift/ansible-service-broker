#! /usr/bin/env python

import sys
import base64
import subprocess

# Output some nicer errors if a user doesn't have the required packages
try:
    import yaml
except Exception:
    print("No yaml parsing modules installed, try: pip install pyyaml")
    sys.exit(1)

# Work around python2/3 input differences
try:
    input = raw_input
except NameError:
    pass

USAGE = """USAGE:
  {command} NAME NAMESPACE IMAGE [BROKER_NAME] [KEY=VALUE]* [@FILE]*

  NAME:         the name of the secret to create/replace
  NAMESPACE:    the target namespace of the secret. It should be the namespace of the broker for most usecases
  IMAGE:        the docker image you would like to associate with the secret
  BROKER_NAME:  the name of the k8s ServiceBroker resource. Defaults to ansible-service-broker
  KEY:          a key to create inside the secret. This cannot contain an "=" sign
  VALUE:        the value for the  KEY in the secret
  FILE:         a yaml loadable file containing key: value pairs. A file must begin with an "@" symbol to be loaded


EXAMPLE:
  {command} mysecret ansible-service-broker docker.io/ansibleplaybookbundle/hello-world-apb key1=hello key2=world @additional_keys.yml

"""

DATA_SEPARATOR = "\n    "

SECRET_TEMPLATE = """---
apiVersion: v1
kind: Secret
metadata:
    name: {name}
    namespace: {namespace}
data:
    {data}
"""


def main():
    name = sys.argv[1]
    namespace = sys.argv[2]
    apb = sys.argv[3]
    if '=' not in sys.argv[4] and '@' not in sys.argv[4]:
        broker_name = sys.argv[4]
        idx = 4
    else:
        broker_name = None
        idx = 3

    keyvalues = list(map(
        lambda x: x.split("=", 1),
        filter(lambda x: "=" in x, sys.argv[idx:])
    ))
    files = list(filter(lambda x: x.startswith("@"), sys.argv[idx:]))
    data = keyvalues + parse_files(files)

    create_secret(name, namespace, data)


def parse_files(files):
    params = []
    for file in files:
        file_name = file[1:]
        with open(file_name, 'r') as f:
            params.extend(yaml.load(f.read()).items())
    return params


def create_secret(name, namespace, data):
    encoded = [(quote(k), base64.b64encode(quote(v))) for (k, v) in data]
    secret = SECRET_TEMPLATE.format(
        name=name,
        namespace=namespace,
        data=DATA_SEPARATOR.join(map(": ".join, encoded))
    )

    with open('/tmp/{name}-secret'.format(name=name), 'w') as f:
        f.write(secret)

    print('oc create -f /tmp/{name}-secret'.format(name=name))

    print('Created secret: \n\n{}'.format(secret))


def quote(string):
    return '"{}"'.format(string)

if __name__ == '__main__':
    if len(sys.argv) < 5 or sys.argv[1] in ("-h", "--help"):
        print(USAGE.format(command=sys.argv[0]))
        sys.exit()

    try:
        main()
    except Exception:
        print("Invalid invocation")
        print(USAGE.format(command=sys.argv[0]))
        raise
