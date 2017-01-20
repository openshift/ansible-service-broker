#!/bin/bash -e

rm -rf vendor
mkdir -p vendor/github.com/openshift/origin
for i in origin/pkg; do ln -s ../../../../$i vendor/github.com/openshift/origin; done
for i in origin/vendor/github.com/openshift/*; do ln -s ../../../$i vendor/github.com/openshift; done
for i in origin/vendor/github.com/*; do [[ $i == */openshift ]] || ln -s ../../$i vendor/github.com; done
for i in origin/vendor/*; do [[ $i == */github.com ]] || ln -s ../$i vendor; done
