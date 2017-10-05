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

try:
    import requests
except Exception:
    print("requests module not installed, try: pip install requests")
    sys.exit(1)

try:
    from apb.engine import broker_request
except Exception:
    print("apb module not installed, try: pip install apb")
    sys.exit(1)


# Work around python2/3 input differences
try:
    input = raw_input
except NameError:
    pass

USAGE = """USAGE:
  {command} NAME NAMESPACE IMAGE [KEY=VALUE]* [@FILE]*

  NAME:      the name of the secret to create/replace
  NAMESPACE: the target namespace of the secret. It should be the namespace of the broker for most usecases
  IMAGE:     the docker image you would like to associate with the secret
  KEY:       a key to create inside the secret. This cannot contain an "=" sign
  VALUE:     the value for the  KEY in the secret
  FILE:      a yaml loadable file containing key: value pairs. A file must begin with an "@" symbol to be loaded


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
    keyvalues = list(map(
        lambda x: x.split("=", 1),
        filter(lambda x: "=" in x, sys.argv[3:])
    ))
    files = list(filter(lambda x: x.startswith("@"), sys.argv[3:]))
    data = keyvalues + parse_files(files)

    runcmd('oc project {}'.format(namespace))
    try:
        runcmd('oc get dc asb')
    except Exception:
        raise Exception("Error: No broker deployment found in namespace {}".format(namespace))
    create_secret(name, namespace, data)
    changed = update_config(name, apb)
    if changed:
        print("Rolling out a new broker...")
        runcmd('oc rollout latest asb')


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

    try:
        runcmd('oc create -f /tmp/{name}-secret'.format(name=name))
    except Exception:
        runcmd('oc replace -f /tmp/{name}-secret'.format(name=name))

    print('Created secret: \n\n{}'.format(secret))


def quote(string):
    return '"{}"'.format(string)


def update_config(name, apb):
    config = get_broker_config()
    secret_entry = {"secret": name, "apb_name": fqname(apb, config), "title": name}
    if secret_entry not in config['data']['broker-config'].get('secrets', []):
        config['data']['broker-config']['secrets'] = config['data']['broker-config'].get('secrets', []) + [secret_entry]
        config_s = format_config(config)
        with open('/tmp/broker-config', 'w') as f:
            f.write(config_s)
        runcmd('oc replace  -f /tmp/broker-config'.format(name=name))
        print('Updated broker config to \n\n{}'.format(config_s))
        return True
    else:
        print("Skipping update to broker configuration becuase secret entry was already present")
        return False


def format_config(config):
    config['data']['broker-config'] = yaml.dump(config['data']['broker-config'])
    for key in ('creationTimestamp', 'resourceVersion', 'selfLink', 'uid'):
        del config['metadata'][key]
    return yaml.dump(config)


def broker_auth(config):
    credentials = {'basic_auth_username': None, 'basic_auth_password': None}
    auth_settings = config['data']['broker-config']['broker'].get('auth')[0]
    if auth_settings.get('type') == 'basic' and auth_settings.get('enabled'):
        secret = yaml.load(runcmd('oc get secret asb-auth-secret -o yaml'))
        credentials = {
            "basic_auth_{}".format(k): base64.b64decode(v)
            for (k, v) in secret['data'].items()
        }
    return credentials


def get_all_apbs(config):
    response = broker_request(None, "/v2/catalog", "get", verify=False, **broker_auth(config))
    return response.json()['services']


def fqname(apb, config):
    search_pattern = apb.split('/')[-1].split(':')[0]
    candidates = get_all_apbs(config)
    matches = [
        str(candidate['name']) for candidate in candidates
        if search_pattern in candidate['name']
    ]

    if not matches:
        print("ERROR: No matches found for {}".format(apb))
        print("apbs found: \n\t- {}".format('\n\t- '.join(
            map(lambda x: x['name'], candidates))
        ))
        sys.exit(1)
    elif len(matches) > 1:
        print("Multiple apbs match...\n")
        for i, match in enumerate(matches):
            print('{}: {}'.format(i+1, match))
        choice = int(input("\nWhich apb would you like to associate?: ")) - 1
        match = matches[choice]
    else:
        match = matches[0]
        print("Associating secret with {}".format(match))
    return match


def get_broker_config():
    config = yaml.load(runcmd("oc get configmap broker-config -o yaml"))
    config['data']['broker-config'] = yaml.load(config['data']['broker-config'])
    return config


def runcmd(cmd):
    print("Running: {}".format(cmd))
    return subprocess.check_output(cmd.split())


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
