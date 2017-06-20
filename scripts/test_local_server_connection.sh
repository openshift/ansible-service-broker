CERT=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt
TOKEN=`cat /var/run/secrets/kubernetes.io/serviceaccount/token`
export KUBERNETES_SERVICE_HOST=192.168.37.1
export KUBERNETES_SERVICE_PORT=8443

curl --cacert ${CERT} -H "Authorization: Bearer ${TOKEN}" https://${KUBERNETES_SERVICE_HOST}:${KUBERNETES_SERVICE_PORT}/version