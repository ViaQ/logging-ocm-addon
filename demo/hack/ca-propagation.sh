#/bin/bash

set -e -u -o pipefail

oc get secret $1 -n $1 -o json | jq -r '.data["ca.crt"]' | base64 -d > /tmp/$1-ca-crt.txt
oc create cm -n openshift-logging $1 --from-file=service-ca.crt=/tmp/$1-ca-crt.txt 

CONFIGMAP_NAME=local-cluster
NAMESPACE=local-cluster
# Create the secret if it doesn't exist
if ! kubectl get configmap "$CONFIGMAP_NAME" -n openshift-logging >/dev/null 2>&1; then
  # Create the secret using the files
  kubectl get secret "$CONFIGMAP_NAME" -n "$NAMESPACE" -o json | jq -r '.data["ca.crt"]' | base64 -d > /tmp/local-cluster-ca-crt.txt
  kubectl create configmap "$CONFIGMAP_NAME" -n openshift-logging --from-file=service-ca.crt=/tmp/local-cluster-ca-crt.txt 
fi
