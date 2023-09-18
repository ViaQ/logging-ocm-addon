# Demo

All steps are meant to be run on the hub cluster except when explicitely said.

### Set up OCP and RHACM

For this demo you will need at least two AWS OCP clusters with one of them (the hub) having at least machines of flavor `m6a.4xlarge` in order to have enough resources for Loki. You will also need to have a S3 Bucket in the same region as the hub cluster.
 
1. Create two OCP clusters
1. Install the Red Hat Advanced Cluster Management operator.
1. Run `oc create ns openshift-logging && oc create ns openshift-operators-redhat`.
1. Create a `MultiClusterHub` resource.
1. *Until Logging 5.8 is released* Deploy Loki Operator manually from main using `make olm-deploy REGISTRY_BASE=quay.io/jmarcal "VERSION=0.1.0-$(git rev-parse --short HEAD)" VARIANT=openshift`
1. Import each spoke cluster RHACM via the web console, using the commands option. Run the commands on the each spoke cluster

### Configure multi cluster on the hub cluster

1. Set the necessary values in `demo/multi-cluster-logging/values.yaml` and deploy `helm upgrade --install mcl demo/multi-cluster-logging/`. This Helm chart will bootstrap configuration on the hub cluster to enabled it to receive logs from the spoke clustes.
1. Run `oc label --overwrite managedcluster/local-cluster cluster.open-cluster-management.io/clusterset=hub-logging-clusters` to label the local-cluster so that the policy applies to it.

## AddOn installation
1. Deploy the addon controller by running `oc apply -k deploy/`.
1. Set the values in `demo/addon-install/values.yaml` and then deploy it with `helm upgrade --install addon-install demo/addon-install/`. This chart will deploy the `ManagedClusterAddOn` to install the AddOn on the spoke clusters and it will deploy the `AddOnDeploymentConfig` resource to configure the AddOn for each spoke cluster. If `certManagerCerts` is enabled it will also create a ConfigMap in openshift-logging with CA bundle of each spoke cluster.