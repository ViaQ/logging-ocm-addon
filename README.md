# Logging OCM AddOn

## Description

The logging-ocm-addon is a pluggable addon working on OCM
rebased on the extensibility provided by
[addon-framework](https://github.com/open-cluster-management-io/addon-framework)
which automates the installation of [cluster-logging-operator](https://github.com/openshift/cluster-logging-operator) and configuration of
[ClusterLogForwarder](https://github.com/openshift/cluster-logging-operator)
on managed clusters to forward logs to a central log store.

The logging-ocm-addon consists of one component:

- **Addon-Manager**: Not only manages the installation of the AddOn on spoke clusters. But also builds the manifests that will be deployed to the spoke clusters.

## Demo

To help demonstrate how the add on can be used we have prepared a [demo](demo/README.md) that goes through all the necesssary steps, from provisioning clusters, to validating that the managed clusters are sending logs.

## Getting started

### Prerequisite

- OCM registration (>= 0.5.0)

### Steps

#### Installing via Kustomize

1. Install the AddOn using Kustomize

```shell
$ kubectl apply -k deploy/
```

2. The addon should now be installed in you hub cluster 
```shell
$ kubectl get ClusterManagementAddOn logging-ocm-addon
```

3. The addon can now be installed it managed clusters by creating `ManagedClusterAddOn` resources in their respective namespaces

## References

- Addon-Framework: [https://github.com/open-cluster-management-io/addon-framework](https://github.com/open-cluster-management-io/addon-framework)