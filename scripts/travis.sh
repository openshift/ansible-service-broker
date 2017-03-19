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

