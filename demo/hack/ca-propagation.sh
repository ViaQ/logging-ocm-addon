#/bin/bash

set -e -u -o pipefail

oc get secret $1 -n $1 -o json | jq -r '.data["ca.crt"]' | base64 -d > /tmp/$1-ca-crt.txt
oc create cm -n openshift-logging $1 --from-file=service-ca.crt=/tmp/$1-ca-crt.txt 