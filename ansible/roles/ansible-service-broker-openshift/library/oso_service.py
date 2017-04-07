#!/usr/bin/python
#
# Copyright 2016 Ansible by Red Hat
#
# This file is part of ansible-container
#

DOCUMENTATION = '''

module: oso_service

short_description: Create or remove a service on a Kubernetes or OpenShift cluster.

description:
  - Create or remove a service on a Kubernetes or OpenShift cluster by setting the C(state) to I(present) or I(absent).
  - The module is idempotent and will not replace an existing service unless the C(reload) option is passed.
  - Supports check mode. Use check mode to view a list of actions the module will take.

options:

'''

EXAMPLES = '''
'''

RETURN = '''
'''
import logging
import logging.config

from ansible.module_utils.basic import *

logger = logging.getLogger('oso_service')

LOGGING = (
    {
        'version': 1,
        'disable_existing_loggers': True,
        'handlers': {
            'console': {
                'level': 'DEBUG',
                'class': 'logging.StreamHandler',
            },
            'file': {
                'level': 'DEBUG',
                'class': 'logging.FileHandler',
                'filename': 'ansible-container.log'
            }
        },
        'loggers': {
            'oso_service': {
                'handlers': ['file'],
                'level': 'INFO',
            },
            'container': {
                'handlers': ['file'],
                'level': 'INFO',
            },
            'compose': {
                'handlers': [],
                'level': 'INFO'
            },
            'docker': {
                'handlers': [],
                'level': 'INFO'
            }
        },
    }
)


class OSOServiceManager(object):

    def __init__(self):

        self.arg_spec = dict(
            project_name=dict(type='str', aliases=['namespace'], required=True),
            state=dict(type='str', choices=['present', 'absent'], default='present'),
            labels=dict(type='dict'),
            ports=dict(type='list', required=True),
            service_name=dict(type='str', required=True),
            loadbalancer=dict(type='bool', default=False),
            replace=dict(type='bool', default=False),
            selector=dict(type='dict', required=True),
            debug=dict(type='bool', default=False)
        )

        self.module = AnsibleModule(self.arg_spec,
                                    supports_check_mode=True)

        self.project_name = None
        self.state = None
        self.labels = None
        self.ports = None
        self.service_name = None
        self.loadbalancer = None
        self.selector = None
        self.replace = None
        self.api = None
        self.debug = None
        self.check_mode = self.module.check_mode

    def exec_module(self):

        for key in self.arg_spec:
            setattr(self, key, self.module.params.get(key))

        if self.debug:
            LOGGING['loggers']['container']['level'] = 'DEBUG'
            LOGGING['loggers']['oso_service']['level'] = 'DEBUG'
        logging.config.dictConfig(LOGGING)

        self.api = OriginAPI(self.module)

        actions = []
        changed = False
        services = dict()
        results = dict()

        project_switch = self.api.set_project(self.project_name)
        if not project_switch:
            actions.append("Create project %s" % self.project_name)
            if not self.check_mode:
                self.api.create_project(self.project_name)

        if self.state == 'present':
            service = self.api.get_resource('service', self.service_name)
            if not service:
                template = self._create_template()
                changed = True
                actions.append("Create service %s" % self.service_name)
                if not self.check_mode:
                    self.api.create_from_template(template=template)
            elif service and self.replace:
                template = self._create_template()
                changed = True
                actions.append("Replace service %s" % self.service_name)
                if not self.check_mode:
                    self.api.replace_from_template(template=template)
            services[self.service_name.replace('-', '_') + '_service'] = self.api.get_resource('service', self.service_name)
        elif self.state == 'absent':
            if self.api.get_resource('service', self.service_name):
                changed = True
                actions.append("Delete service %s" % self.service_name)
                if not self.check_mode:
                    self.api.delete_resource('service', self.service_name)

        results['changed'] = changed

        if self.check_mode:
            results['actions'] = actions

        if services:
            results['ansible_facts'] = services

        self.module.exit_json(**results)

    def _create_template(self):
        '''
        apiVersion: v1
            kind: Service
            metadata:
              name: frontend
              labels:
                app: guestbook
                tier: frontend
            spec:
              # if your cluster supports it, uncomment the following to automatically create
              # an external load-balanced IP for the frontend service.
              # type: LoadBalancer
              ports:
                # the port that this service should serve on
              - port: 80
              selector:
                app: guestbook
                tier: frontend
        '''

        selector = self.selector if self.selector else self.service_name
        self._update_ports()

        template = dict(
            apiVersion="v1",
            kind="Service",
            metadata=dict(
                name=self.service_name
            ),
            spec=dict(
                selector=selector,
                ports=self.ports
            )
        )

        if self.labels:
            template['metadata']['labels'] = self.labels

        if self.loadbalancer:
            template['spec']['type'] = 'LoadBalancer'

        return template

    def _update_ports(self):
        count = 0
        for port in self.ports:
            if not port.get('name'):
                port['name'] = "port%s" % count
                count += 1


#The following will be included by `ansble-container shipit` when cloud modules are copied into the role library path.

import re
import json


class OriginAPI(object):

    def __init__(self, module, target="oc"):
        self.target = target
        self.module = module

    @staticmethod
    def use_multiple_deployments(services):
        '''
        Inspect services and return True if the app supports multiple replica sets.

        :param services: list of docker-compose service dicts
        :return: bool
        '''
        multiple = True
        for service in services:
            if not service.get('ports'):
                multiple = False
            if service.get('volumes_from'):
                multiple = False
        return multiple

    def call_api(self, cmd, data=None, check_rc=False, error_msg=None):
        rc, stdout, stderr = self.module.run_command(cmd, data=data)
        logger.debug("Received rc: %s" % rc)
        logger.debug("stdout:")
        logger.debug(stdout)
        logger.debug("stderr:")
        logger.debug(stderr)

        if check_rc and rc != 0:
            self.module.fail_json(msg=error_msg, stderr=stderr, stdout=stdout)

        return rc, stdout, stderr

    def create_from_template(self, template=None, template_path=None):
        if template_path:
            logger.debug("Create from template %s" % template_path)
            error_msg = "Error Creating %s" % template_path
            cmd = "%s create -f %s" % (self.target, template_path)
            rc, stdout, stderr = self.call_api(cmd, check_rc=True, error_msg=error_msg)
            return stdout

        if template:
            logger.debug("Create from template:")
            formatted_template = json.dumps(template, sort_keys=False, indent=4, separators=(',', ':'))
            logger.debug(formatted_template)
            cmd = "%s create -f -" % self.target
            rc, stdout, stderr = self.call_api(cmd, data=formatted_template, check_rc=True,
                                               error_msg="Error creating from template.")
            return stdout

    def replace_from_template(self, template=None, template_path=None):
        if template_path:
            logger.debug("Replace from template %s" % template_path)
            cmd = "%s replace -f %s" % (self.target, template_path)
            error_msg = "Error replacing %s" % template_path
            rc, stdout, stderr = self.call_api(cmd, check_rc=True, error_msg=error_msg)
            return stdout
        if template:
            logger.debug("Replace from template:")
            formatted_template = json.dumps(template, sort_keys=False, indent=4, separators=(',', ':'))
            logger.debug(formatted_template)
            cmd = "%s replace -f -" % self.target
            rc, stdout, stderr = self.call_api(cmd, data=formatted_template, check_rc=True,
                                               error_msg="Error replacing from template")
            return stdout

    def delete_resource(self, type, name):
        cmd = "%s delete %s/%s" % (self.target, type, name)
        logger.debug("exec: %s" % cmd)
        error_msg = "Error deleting %s/%s" % (type, name)
        rc, stdout, stderr = self.call_api(cmd, check_rc=True, error_msg=error_msg)
        return stdout

    def get_resource(self, type, name):
        result = None
        cmd = "%s get %s/%s -o json" % (self.target, type, name)
        logger.debug("exec: %s" % cmd)
        rc, stdout, stderr = self.call_api(cmd)
        if rc == 0:
            result = json.loads(stdout) 
        elif rc != 0 and not re.search('not found', stderr):
            error_msg = "Error getting %s/%s" % (type, name)
            self.module.fail_json(msg=error_msg, stderr=stderr, stdout=stdout)
        return result
   
    def set_context(self, context_name):
        cmd = "%s user-context %s" % (self.target, context_name)
        logger.debug("exec: %s" % cmd)
        error_msg = "Error switching to context %s" % context_name
        rc, stdout, stderr = self.call_api(cmd, check_rc=True, error_msg=error_msg)
        return stdout

    def set_project(self, project_name):
        result = True
        cmd = "%s project %s" % (self.target, project_name)
        logger.debug("exec: %s" % cmd)
        rc, stdout, stderr = self.call_api(cmd)
        if rc != 0:
            result = False
            if not re.search('does not exist', stderr):
                error_msg = "Error switching to project %s" % project_name
                self.module.fail_json(msg=error_msg, stderr=stderr, stdout=stdout)
        return result

    def create_project(self, project_name):
        result = True
        cmd = "%s new-project %s" % (self.target, project_name)
        logger.debug("exec: %s" % cmd)
        error_msg = "Error creating project %s" % project_name
        self.call_api(cmd, check_rc=True, error_msg=error_msg)
        return result

    def get_deployment(self, deployment_name):
        cmd = "%s deploy %s" % (self.target, deployment_name)
        logger.debug("exec: %s" % cmd)
        rc, stdout, stderr = self.call_api(cmd)
        if rc != 0:
            if not re.search('not found', stderr):
                error_msg = "Error getting deployment state %s" % deployment_name
                self.module.fail_json(msg=error_msg, stderr=stderr, stdout=stdout)
        return stdout


def main():
    manager = OSOServiceManager()
    manager.exec_module()

if __name__ == '__main__':
    main()
