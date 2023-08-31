#/bin/bash

set -e -u -o pipefail

kubectl get secret local-cluster -n local-cluster -o json | jq -r '.data["tls.crt"]' | base64 -d > /tmp/local-cluster-crt.txt
kubectl get secret local-cluster -n local-cluster -o json | jq -r '.data["tls.key"]' | base64 -d > /tmp/local-cluster-key.txt
kubectl get cm -n openshift-logging lokistack-hub-gateway-ca-bundle -o json | jq -r '.data["service-ca.crt"]' > /tmp/loki-ca-bundle.txt

SECRET_NAME=mtls-spoke-hub
NAMESPACE=openshift-logging

# Check if the secret doesn't exist
if kubectl get secret "$SECRET_NAME" -n "$NAMESPACE" >/dev/null 2>&1; then
  kubectl delete secret "$SECRET_NAME" -n "$NAMESPACE"
  sleep 3
fi
# Create the secret using the files
kubectl create secret generic "$SECRET_NAME" -n "$NAMESPACE" --from-file=ca-bundle.crt=/tmp/loki-ca-bundle.txt --from-file=tls.crt=/tmp/local-cluster-crt.txt --from-file=tls.key=/tmp/local-cluster-key.txt