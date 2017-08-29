#! /usr/bin/env python
import sys
import yaml
import subprocess

USAGE = """USAGE:
  secrets.py NAME NAMESPACE IMAGE KEY=VALUE [KEY=VALUE]*
"""

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
    data = list(map(lambda x: x.split("="), filter(lambda x: "=" in x, sys.argv[3:])))

    runcmd('oc project {}'.format(namespace))
    try:
        runcmd('oc get dc asb')
    except Exception:
        raise Exception("Error: No broker deployment found in namespace {}".format(namespace))
    create_secret(name, namespace, data)
    update_config(name, apb)
    print("Rolling out a new broker...")
    runcmd('oc rollout latest asb')



def create_secret(name, namespace, data):
    secret = SECRET_TEMPLATE.format(
        name=name,
        namespace=namespace,
        data="\n  ".join(map(lambda x: ": ".join(map(quote, x)), data))
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
    secret_entry = [{"secret" : name, "apb_name": fqname(apb), "title": name}]
    config = get_broker_config()
    config['data']['broker-config']['secrets'] = config.get('secrets', []) + secret_entry
    config_s = format_config(config)
    with open('/tmp/broker-config', 'w') as f:
        f.write(config_s)
    runcmd('oc replace  -f /tmp/broker-config'.format(name=name))

    print('Updated broker config to \n\n{}'.format(config_s))


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
        print(USAGE)
        sys.exit()

    try:
        main()
    except Exception:
        print("Invalid invocation")
        print(USAGE)
        raise
