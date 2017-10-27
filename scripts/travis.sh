#!/bin/bash

#
# travis.sh
#
# This script is used by travis to test the broker
#
action=$1

export GLIDE_TARBALL="https://github.com/Masterminds/glide/releases/download/v0.12.3/glide-v0.12.3-linux-amd64.tar.gz"

if [[ "$action" == "before_install" ]]; then
  echo "================================="
  echo "        Before Install           "
  echo "================================="
  sudo do-release-upgrade -f DistUpgradeViewNonInteractive
  sudo apt-get -qq update
  sudo apt-get install -y python-apt autoconf pkg-config e2fslibs-dev libblkid-dev zlib1g-dev liblzo2-dev asciidoc
elif [[ "$action" == "install" ]]; then
  echo "================================="
  echo "           Install               "
  echo "================================="
  # Install ansible
  sudo pip install ansible==2.3.1 pyOpenSSL==16.2.0

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

  # build btrfs from scratch
  git clone https://github.com/kdave/btrfs-progs.git
  cd btrfs-progs
  ./autogen.sh
  ./configure
  make
  sudo make install
  cd $TRAVIS_BUILD_DIR

  # build nsenter from scratch
  mkdir util-linux
  wget https://www.kernel.org/pub/linux/utils/util-linux/v2.24/util-linux-2.24.1.tar.gz -qO - | tar -xz -C util-linux
  sudo apt-get install libncurses5-dev libslang2-dev gettext zlib1g-dev libselinux1-dev debhelper lsb-release pkg-config po-debconf autoconf automake autopoint libtool
  cd util-linux/util-linux-2.24.1
  ./autogen.sh
  ./configure && make
  sudo cp ./nsenter /usr/bin
  cd $TRAVIS_BUILD_DIR

  # now install deps
  wget -O /tmp/glide.tar.gz $GLIDE_TARBALL
  tar xfv /tmp/glide.tar.gz -C /tmp
  sudo mv $(find /tmp -name "glide") /usr/bin

  # install golint
  go get -u github.com/golang/lint/golint
elif [[ "$action" == "before_script" ]]; then
  echo "================================="
  echo "          Before Script          "
  echo "================================="
  sudo ufw disable
  tmp=`mktemp`
  echo '{"insecure-registries":["172.30.0.0/16"]}' > ${tmp}
  sudo mv ${tmp} /etc/docker/daemon.json
  sudo mount --make-shared /
  sudo service docker restart
elif [[ "$action" == "lint" ]]; then
  echo "================================="
  echo "              Lint               "
  echo "================================="
  make lint
elif [[ "$action" == "format" ]]; then
  echo "================================="
  echo "             Format              "
  echo "================================="
  make fmtcheck
elif [[ "$action" == "vet" ]]; then
  echo "================================="
  echo "              Vet                "
  echo "================================="
  make vet
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
  if [ $? -eq 0 ]; then
      make ci LOCAL_CI=false
  fi
fi
