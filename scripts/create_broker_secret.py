#! /usr/bin/env python
import sys
import yaml
import subprocess

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

DATA_SEPARATOR="\n    "

SECRET_TEMPLATE = """---
apiVersion: v1
kind: Secret
metadata:
    name: {name}
    namespace: {namespace}
stringData:
    {data}
"""

def main():
    name = sys.argv[1]
    namespace = sys.argv[2]
    apb = sys.argv[3]
    keyvalues = list(map(lambda x: x.split("=", 1), filter(lambda x: "=" in x, sys.argv[3:])))
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
    for file  in files:
        file_name = file[1:]
        with open(file_name, 'r') as f:
            params.extend(yaml.load(f.read()).items())
    return params


def create_secret(name, namespace, data):
    secret = SECRET_TEMPLATE.format(
        name=name,
        namespace=namespace,
        data=DATA_SEPARATOR.join(map(lambda x: ": ".join(map(quote, x)), data))
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
    secret_entry = {"secret" : name, "apb_name": fqname(apb), "title": name}
    config = get_broker_config()
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

def fqname(apb):
    registries = {'docker.io': 'dh'}
    registry, org, end = apb.split('/')

    if ":" in end:
        image, tag = end.split(":")
    else:
        image = end
        tag = 'latest'

    return '-'.join([registries[registry], org, image, tag])


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
