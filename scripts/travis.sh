#!/bin/bash

#
# travis.sh
#
# This script is used by travis to test the broker
#
action=$1

export GLIDE_TARBALL="https://github.com/Masterminds/glide/releases/download/v0.12.3/glide-v0.12.3-linux-amd64.tar.gz"
export PROJECT_ROOT=$GOPATH/src/github.com/fusor/ansible-service-broker

if [[ "$action" == "install" ]]; then

  # HACK stolen from this PR
  # https://github.com/pnasrat/docker/commit/6a18493259c3f201dc28fe114e592b0f14c725d1
  sudo git clone https://git.fedorahosted.org/git/lvm2.git /usr/local/lvm2 && cd /usr/local/lvm2 && git checkout v2_02_103
  sudo cd /usr/local/lvm2 && ./configure --enable-static-link && make device-mapper && make install_device-mapper

  # return to our normal install procedures
  wget -O /tmp/glide.tar.gz $GLIDE_TARBALL
  tar xfv /tmp/glide.tar.gz -C /tmp
  sudo mv $(find /tmp -name "glide") /usr/bin
  cd $PROJECT_ROOT && glide install
elif [[ "$action" == "lint" ]]; then
  echo "================================="
  echo "             Lint                "
  echo "================================="
  CMD_PASS=$(gofmt -d $PROJECT_ROOT/cmd 2>&1 | read; echo $?)
  PKG_PASS=$(gofmt -d $PROJECT_ROOT/pkg 2>&1 | read; echo $?)
  echo "CMD_PASS=$CMD_PASS"
  echo "PKG_PASS=$PKG_PASS"
  FULL_PASS=$([[ $CMD_PASS == 1 ]] && [[ $PKG_PASS == 1 ]]; echo $?)
  echo "FULL_PASS=$FULL_PASS"
  echo "================================="
  exit $FULL_PASS
elif [[ "$action" == "build" ]]; then
  echo "================================="
  echo "             Build               "
  echo "================================="
  make build
  exit $?
elif [[ "$action" == "test" ]]; then
  echo "================================="
  echo "             Test                "
  echo "================================="
  make test
fi

