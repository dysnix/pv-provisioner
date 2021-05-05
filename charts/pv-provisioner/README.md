# Persistent Volume Provisioner

## Introduction

This chart bootstraps a Persistent Volume Provisioner from Snapshots.
You can get more information about pv-provisioner by https://git.dysnix.com/dysnix/pv-provisioner

## Prerequisites

* Google Cloud Platform
* Kubernetes
* Helm

## Installing the Chart

1. Create GCP Service Account and save as `gcp-cred.json` file in root dir for helm chart

2. Set variables in `values.yaml` file

3. Deploy helm chart

```console
helm install --name pv-provisioner ./
```

## Configuration

The following table lists the configurable parameters of the vault chart and their default values.

| Parameter                         | Description                                   | Default                               |
|-----------------------------------|-----------------------------------------------|---------------------------------------|
| `GCP_PROJECT`                     | Google Cloud Platform Project ID              |                                       |
| `GCP_ZONE`                        | Availability zone of disks                    |                                       |
| `GCP_SNAPSHOT_LABEL`              | Kubernetes Label name for search              | `app`                                 |

