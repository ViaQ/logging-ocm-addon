# Demo

All steps are meant to be run on the hub cluster except when explicitely said.

1. Create two OCP clusters
1. Install the Red Hat Advanced Cluster Management operator.
1. Deploy the log-storage-toolbox/bootstrap.
1. Create a `MultiClusterHub` resource.
1. *Until 5.8 is released* Deploy Loki Operator manually from main using `make olm-deploy REGISTRY_BASE=quay.io/jmarcal "VERSION=0.1.0-$(git rev-parse --short HEAD)" VARIANT=openshift`
1. Import each spoke cluster RHACM via the web console, using the commands option. Run the commands on the each spoke cluster
1. Set the necessary parameters in `demo/rhacm-bootstrap/overlays/production-hub` and deploy `oc apply -k demo/rhacm-bootstrap/overlays/production-hub/`. This Kustomize will bootstrap the RHACM configuration.
1. Label the clusters to add them to the respective clustersets.
1. Build and push the addon with `make oci`, don't forget to set the registry with `export REGISTRY_BASE=quay.io/MY_REGISRTY`
1. Deploy the addon by running `oc apply -k deploy/`
1. Set the values for the Helm chart under `demo/logging-addon-clusters-resources` and then deploy it with `helm upgrade --install test demo/logging-addon-clusters-resources/`. This chart will deploy the `ManagedClusterAddOn` to install the AddOn on the spoke clusters and it will deploy the `AddOnDeploymentConfig` resource to configure the AddOn for each cluster. 
1. Run the script `./demo/hack/mtls-secret-local.sh` to create the mtls-spoke-hub secret on the hub cluster
1. Run the script `./demo/hack/ca-propagation.sh CLUSTER_NAME` to create the CAs configMaps in the openshift-logging namespace



### Add a managed cluster to the `hub-logging-clusters` set:

```shell
oc label --overwrite managedcluster/<CLUSTER_NAME> cluster.open-cluster-management.io/clusterset=hub-logging-clusters
```

### Add a managed cluster to the `spoke-logging-clusters` set:

```shell
oc label --overwrite managedcluster/<CLUSTER_NAME> cluster.open-cluster-management.io/clusterset=spoke-logging-clusters
```

