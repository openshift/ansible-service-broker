#!/bin/bash

#
# travis.sh
#
# This script is used by travis to test the broker
#
action=$1

if [[ "$action" == "before_install" ]]; then
  echo "================================="
  echo "        Before Install           "
  echo "================================="
  sudo apt-get install -y python-apt autoconf pkg-config e2fslibs-dev libblkid-dev zlib1g-dev liblzo2-dev asciidoc
elif [[ "$action" == "install" ]]; then
  echo "================================="
  echo "           Install               "
  echo "================================="
  # Install ansible
  sudo pip install ansible pyOpenSSL

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
  cd $TRAVIS_BUILD_DIR

  # now install deps
  go get -u github.com/golang/dep/cmd/dep

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
elif [[ "$action" == "docs" ]]; then
  echo "================================="
  echo "           Doc Change            "
  echo "================================="
  # Copied from https://github.com/facebook/react/pull/2000
  git diff --name-only HEAD^ | grep -qvE '(\.md$)|(^(docs|examples))/' || {
      echo "Only docs were updated, stopping build process."
      exit 9
  }
elif [[ "$action" == "lint" ]]; then
  echo "================================="
  echo "              Lint               "
  echo "================================="
  # install golint
  go get -u github.com/golang/lint/golint

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
  # now install deps
  go get -u github.com/golang/dep/cmd/dep
  make vendor
  make build
  exit $?
elif [[ "$action" == "test" ]]; then
  echo "================================="
  echo "              Test               "
  echo "================================="
  # install goveralls for coveralls integration
  go get github.com/mattn/goveralls

  make ci-test-coverage
elif [[ "$action" == "setup-cluster" ]]; then
  echo "================================="
  echo "          Setup Cluster          "
  echo "================================="
  # Add an arguemnt only when running Travis for the broker
  ./scripts/broker-ci/setup-cluster.sh broker
  exit $?
elif [[ "$action" == "setup-broker" ]]; then
  echo "================================="
  echo "          Setup Broker           "
  echo "   (Only for Broker Travis job)  "
  echo "================================="
  ./scripts/broker-ci/setup-broker.sh
  exit $?
elif [[ "$action" == "ci" ]]; then
  echo "================================="
  echo "            Broker CI            "
  echo "================================="
  make ci
  exit $?
elif [[ "$action" == "k8s-ci" ]]; then
  echo "================================="
  echo "    Broker CI for Kubernetes     "
  echo "================================="
  make ci-k
  exit $?
elif [[ "$action" == "pv-setup" ]]; then
  ./scripts/broker-ci/pv-setup.sh
  exit $?
elif [[ "$action" == "setup-minikube" ]]; then
  ./scripts/broker-ci/setup-minikube.sh
  exit $?
elif [[ "$action" == "k8s-catalog" ]]; then
  ./scripts/broker-ci/setup-catalog.sh
  exit $?
elif [[ "$action" == "k8s-broker" ]]; then
  TAG=canary ./scripts/run_latest_k8s_build.sh
  exit $?
fi
