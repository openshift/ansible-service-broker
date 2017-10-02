#!/bin/bash

######
# Main
######

# Test with latest code first. Then, we'll move to a release branch
git clone https://github.com/rthallisey/service-broker-ci
pushd service-broker-ci
make run
popd

# pushd vendor/github.com/rthallisey/service-broker-ci
# make run
# popd
