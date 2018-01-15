#/bin/bash

set -e

function generate_pv() {
  local basedir="${1}"
  local name="${2}"

  cat <<EOF
apiVersion: v1
kind: PersistentVolume
metadata:
  name: ${name}
  labels:
    volume: ${name}
spec:
  capacity:
    storage: 100Gi
  accessModes:
    - ReadWriteOnce
    - ReadWriteMany
    - ReadOnlyMany
  hostPath:
    path: ${basedir}/${name}
  persistentVolumeReclaimPolicy: Recycle
EOF
}

function setup_pv_dir() {
  local dir="${1}"
  if [[ ! -d "${dir}" ]]; then
    sudo mkdir -p "${dir}"
  fi
  if ! chcon -t svirt_sandbox_file_t "${dir}" &> /dev/null; then
    echo "Not setting SELinux content for ${dir}"
  fi
  sudo chmod 777 "${dir}"
  sudo chown $(whoami): "${dir}"
}

function create_pv() {
  local basedir="${1}"
  local name="${2}"
  setup_pv_dir "${basedir}/${name}"
  if ! kubectl get pv "${name}" &> /dev/null; then
    generate_pv "${basedir}" "${name}" | kubectl create -f -
  else
    echo "persistentvolume ${name} already exists"
  fi
}

pv="pv${2}"
basedir="${1}"
setup_pv_dir "${basedir}"

create_pv "${basedir}" "${pv}"
