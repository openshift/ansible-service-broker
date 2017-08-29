#!/bin/env python

import os
import jinja2
import yaml

def render(tpl_path, content):
    path, filename = os.path.split(tpl_path)
    return jinja2.Environment(
        loader=jinja2.FileSystemLoader(path or './')
    ).get_template(filename).render(content)

with open('k8s-variables.yaml', 'r') as content_file:
    data = content_file.read()

content = yaml.load(data)

result = render('./k8s-ansible-service-broker.yaml.j2', content)

with open('k8s-ansible-service-broker.yaml', 'w') as rendered_file:
    rendered_file.write(result)
