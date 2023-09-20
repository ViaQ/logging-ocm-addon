# Demo: Multi-cluster log collection and forwarding

The following steps demonstrate how to use the `logging-ocm-addon` to manage `ClusterLogging` and `ClusterLogForwarder` resources across a Red Hat Advanced Cluster Management (RHACM) managed fleet of OpenShift (OCP) clusters. The `logging-ocm-addon` is limited only to manage log collection and forwarding. Thus step 2 is dedicated to install `LokiStack` as the storage solution on the RHACM hub cluster.

All steps are meant to be run on the hub cluster except when explicitely said.

## Prerequisites: RHACM and OCP cluster fleet

For this demo you will need at least two OCP clusters (hosted on AWS) with one of them (the hub) having at least machines of flavor `m6a.4xlarge` in order to have enough resources for `LokiStack`. You will also need to have a S3 Bucket in the same region as the hub cluster.
 
1. Use the OpenShift Installer to create and setup two OCP cluster on AWS.
1. Install the `Red Hat Advanced Cluster Management` operator.
1. Run `oc create ns openshift-logging && oc create ns openshift-operators-redhat`.
1. Create a `MultiClusterHub` resource.
1. Install the `Loki Operator`: *Until Logging 5.8 is released* manually from main using `make olm-deploy REGISTRY_BASE=quay.io/$QUAY_IO_USERNAME "VERSION=0.1.0-$(git rev-parse --short HEAD)" VARIANT=openshift`
1. Import each spoke cluster `RHACM` via the web console, using the commands option by running the commands on each spoke cluster.

## Install LokiStack on the hub cluster

The following steps use Helm to install a set of RHACM `ConfigurationPolicies` that resolve `LokiStack` installation, mTLS-based tenant configuration and placement on the hub cluster. 

_Hint:_ The `certManagerCerts` installs additionally the `CertManager` operator on the hub cluster. It demonstrates the ability to delegate PKI management for all tenants to a third-party tool.

1. Prepare the Helm chart configuration setting the necessary values in `demo/multi-cluster-logging/values.yaml`
1. Deploy the LokiStack and PKI executing `helm upgrade --install mcl demo/multi-cluster-logging/`. This Helm chart will bootstrap configuration on the hub cluster to enabled it to receive logs from the spoke clustes.
1. Run `oc label --overwrite managedcluster/local-cluster cluster.open-cluster-management.io/clusterset=hub-logging-clusters` to label the `local-cluster` a.k.a. hub so that the policy applies to it.

## Install logging-ocm-addon

1. Deploy the addon controller by running `oc apply -k deploy/`.

## Manage Log Collection and Forwarding across spoke sClusters

The following chart will deploy the `ManagedClusterAddOn` resource that installs the AddOn on each spoke cluster. In addition it will deploy the `AddOnDeploymentConfig` resource to configure the AddOn for each spoke cluster. 

_Hint:_ The `certManagerCerts` if enabled will create a `ConfigMap` in `openshift-logging` holding the CA bundle propagated to each spoke cluster.

1. Set the values in `demo/addon-install/values.yaml`.
1. Deploy it with `helm upgrade --install addon-install demo/addon-install/`. 


