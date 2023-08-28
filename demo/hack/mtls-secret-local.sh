#/bin/bash

set -e -u -o pipefail

oc get secret local-cluster -n local-cluster -o json | jq -r '.data["tls.crt"]' | base64 -d > /tmp/local-cluster-crt.txt
oc get secret local-cluster -n local-cluster -o json | jq -r '.data["tls.key"]' | base64 -d > /tmp/local-cluster-key.txt
oc get cm -n openshift-logging lokistack-hub-ca-bundle -o json | jq -r '.data["service-ca.crt"]' > /tmp/loki-ca-bundle.txt
oc create secret -n openshift-logging generic mtls-spoke-hub --from-file=ca-bundle.crt=/tmp/loki-ca-bundle.txt --from-file=tls.crt=/tmp/local-cluster-crt.txt --from-file=tls.key=/tmp/local-cluster-key.txt