#!/bin/bash

set -x

function install-nsenter {
    wget https://www.kernel.org/pub/linux/utils/util-linux/v2.24/util-linux-2.24.1.tar.gz -qO - | tar -xz -C ./
    sudo apt-get install libncurses5-dev libslang2-dev gettext zlib1g-dev libselinux1-dev debhelper lsb-release pkg-config po-debconf autoconf automake autopoint libtool -y

    pushd ./util-linux-2.24.1
    ./autogen.sh
    ./configure && make

    ./nsenter -V
    sudo cp ./nsenter /usr/bin
    popd
}

sudo apt-get install python-jinja2 coreutils util-linux socat -y
install-nsenter

hostname=$(ip addr show docker0 | grep -Po 'inet \K[\d.]+')
sudo iptables -t nat -I POSTROUTING ! -o docker0 -s $hostname/16 -j MASQUERADE
sudo iptables -I FORWARD -o docker0 -j ACCEPT
sudo iptables -I FORWARD -i docker0 ! -o docker0 -j ACCEPT

curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl
curl -Lo minikube https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo chmod +x kubectl && sudo mv kubectl /usr/local/bin/
sudo chmod +x minikube && sudo mv minikube /usr/local/bin/

echo "Starting Minikube"
sudo minikube start --vm-driver=none --extra-config=apiserver.Authorization.Mode=RBAC
minikube update-context

JSONPATH='{range .items[*]}{@.metadata.name}:{range @.status.conditions[*]}{@.type}={@.status};{end}{end}'; until kubectl get nodes -o jsonpath="$JSONPATH" 2>&1 | grep -q "Ready=True"; do sleep 1; done
echo "Minikube started"
