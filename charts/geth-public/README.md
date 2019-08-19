# Ethereum Public Network

[Geth](https://geth.ethereum.org) is official Go implementation of the Ethereum protocol

## Introduction

This chart bootstraps a Statefulset Ethereum nodes cluster on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

## Prerequisites

- Kubernetes 1.8+
- PV provisioner support in the underlying infrastructure (now support only AWS)

## Installing the Chart
To install the chart with the release name `my-release`:

```bash
$ helm install --name my-release stable/geth
```

The command deploys Geth on the Kubernetes cluster in the default configuration.
The [configuration](#configuration) section lists the parameters that can be configured during installation.

> **Tip**: List all releases using `helm list`

## Uninstalling the Chart

To uninstall/delete the `my-release`:

```bash
$ helm delete my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

```bash
$ helm install --name my-release -f values.yaml stable/geth
```

> **Tip**: You can use the default [values.yaml](values.yaml)

## Persistence

The geth image stores the geth node data (Blockchain and wallets) and configurations at the `/root` path of the container.

By default a PersistentVolumeClaim is created and mounted into that directory. In order to disable this functionality
you can change the values.yaml to disable persistence and use an emptyDir instead.

> *"An emptyDir volume is first created when a Pod is assigned to a Node, and exists as long as that Pod is running on that node. When a Pod is removed from a node for any reason, the data in the emptyDir is deleted forever."*

!!! WARNING !!!

Please NOT use emptyDir for production cluster! Your wallets will be lost on container restart!

