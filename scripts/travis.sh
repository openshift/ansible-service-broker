#!/bin/bash

#
# travis.sh
#
# This script is used by travis to test the broker
#
action=$1

export GLIDE_TARBALL="https://github.com/Masterminds/glide/releases/download/v0.12.3/glide-v0.12.3-linux-amd64.tar.gz"
export PROJECT_ROOT=$GOPATH/src/github.com/openshift/ansible-service-broker

if [[ "$action" == "install" ]]; then
  # dash? wtf is dash? UGH! use a real shell
  sudo rm /bin/sh
  sudo ln -s  /bin/bash /bin/sh

  # install devmapper from scratch
  cd $HOME
  git clone http://sourceware.org/git/lvm2.git
  cd lvm2
  ./configure
  sudo make install_device-mapper
  cd ..

  #  build btrfs from scratch
  git clone https://github.com/kdave/btrfs-progs.git
  cd btrfs-progs
  ./autogen.sh
  ./configure
  make
  sudo make install
  cd $PROJECT_ROOT

  # now install deps
  wget -O /tmp/glide.tar.gz $GLIDE_TARBALL
  tar xfv /tmp/glide.tar.gz -C /tmp
  sudo mv $(find /tmp -name "glide") /usr/bin

  # install golint
  go get -u github.com/golang/lint/golint
elif [[ "$action" == "lint" ]]; then
  echo "================================="
  echo "              Lint               "
  echo "================================="
  # to have lint fail build add -set_exit_status option
  # to each golint command
  if [ -n "$(golint $PROJECT_ROOT/cmd/...)" ]; then
        golint $PROJECT_ROOT/cmd/...
        exit 1
  fi
  if [ -n "$(golint $PROJECT_ROOT/pkg/...)" ]; then
        golint $PROJECT_ROOT/pkg/...
        exit 1
  fi
elif [[ "$action" == "format" ]]; then
  echo "================================="
  echo "             Format              "
  echo "================================="
  if [ -n "$(gofmt -l $PROJECT_ROOT/cmd $PROJECT_ROOT/pkg)" ]; then
      gofmt -d $PROJECT_ROOT/cmd $PROJECT_ROOT/pkg
      exit 1
  fi
elif [[ "$action" == "vet" ]]; then
  echo "================================="
  echo "              Vet                "
  echo "================================="
  go tool vet $PROJECT_ROOT/cmd $PROJECT_ROOT/pkg
  exit $?
elif [[ "$action" == "build" ]]; then
  echo "================================="
  echo "             Build               "
  echo "================================="
  make vendor
  make build
  exit $?
elif [[ "$action" == "test" ]]; then
  echo "================================="
  echo "              Test               "
  echo "================================="
  make test
elif [[ "$action" == "ci" ]]; then
  echo "================================="
  echo "            Broker CI            "
  echo "================================="
  ./scripts/broker-ci/setup.sh
#  make ci
fi
