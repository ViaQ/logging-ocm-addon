# Demo

All steps are meant to be run on the hub cluster except when explicitely said.

1. Create two OCP clusters
1. Install the Red Hat Advanced Cluster Management operator.
1. Deploy the log-storage-toolbox/bootstrap.
1. Increase the number of worker nodes by editing the machinesets (necessary for Loki). Each AZ should have 3 nodes. `oc edit machinesets -n openshift-machine-api MACHINESET_NAME`
1. Create a `MultiClusterHub` resource.
1. *Until Logging 5.8 is released* Deploy Loki Operator manually from main using `make olm-deploy REGISTRY_BASE=quay.io/jmarcal "VERSION=0.1.0-$(git rev-parse --short HEAD)" VARIANT=openshift`
1. Import each spoke cluster RHACM via the web console, using the commands option. Run the commands on the each spoke cluster
1. Set the necessary parameters in `demo/rhacm-bootstrap/overlays/production-hub` and deploy `oc apply -k demo/rhacm-bootstrap/overlays/production-hub/`. This Kustomize will bootstrap the RHACM configuration.
1. Label the clusters to [add them to the respective clustersets](#assign-cluster-to-cluster-set). 
1. Build and push the addon with `make oci`, don't forget to set the registry with `export REGISTRY_BASE=quay.io/MY_REGISRTY`
1. Deploy the addon by running `oc apply -k deploy/`, before running the command make sure to point the image to your registry.
1. Set the values for the Helm chart under `demo/logging-addon-clusters-resources` and then deploy it with `helm upgrade --install logging-omc-addon demo/logging-addon-clusters-resources/`. This chart will deploy the `ManagedClusterAddOn` to install the AddOn on the spoke clusters and it will deploy the `AddOnDeploymentConfig` resource to configure the AddOn for each cluster. If `certManagerCerts` is enabled it will also configure a Self sign CA and provision certificates for the clusters and to the local-cluster
1. Run the script `./demo/hack/ca-propagation.sh CLUSTER_NAME` to create the CAs configMaps in the openshift-logging namespace. The first time you run it it will also propagate the local-cluster CA to openshift-monitoring
1. Run the script `./demo/hack/mtls-secret-local.sh` to create the mtls-spoke-hub secret on the hub cluster. If run again the secret will be recreated


## Assign cluster to cluster set

### Add a managed cluster to the `hub-logging-clusters` set:

```shell
oc label --overwrite managedcluster/<CLUSTER_NAME> cluster.open-cluster-management.io/clusterset=hub-logging-clusters
```

### Add a managed cluster to the `spoke-logging-clusters` set:

```shell
oc label --overwrite managedcluster/<CLUSTER_NAME> cluster.open-cluster-management.io/clusterset=spoke-logging-clusters
```

