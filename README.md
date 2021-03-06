# pv-provisioner

Kubernetes Persistent Volume Provisioner from pre-created snapshots.

## How its works?

* You create PersistentVolumeClaim with empty Storage class (for
  disable [Dynamic Provisioning](https://kubernetes.io/docs/concepts/storage/dynamic-provisioning/#enabling-dynamic-provisioning))
* _pv-provisioner_ detect PVC and track tag `app` value
* _pv-provisioner_ search most actual snapshot with label `app` value equal `app` tag of k8s PVC
* _pv-provisioner_ create new Disk/Volume
* _pv-provisioner_ create new PersistentVolume with required params
* PVC use created PV

## Deploy

Please follow [charts/pv-provisioner/README.md](./charts/pv-provisioner/README.md)

## Supported cloud platforms

* "gcp" - Google Cloud Platform
  * if you're using Kubernetes Serviceaccount with `Workload Identity` please set appropriate annotations to the service account in `values.yaml` file.

## GCP Environment variables

* **GCP_PROJECT** - GCP project ID
* **GCP_ZONE** - GCP Disks availability zone
* **GCP_SNAPSHOT_LABEL** - tag name for find Snapshot
* **GCP_DISK_TYPE** - Disk type resource describing which disk type to use to create the disk. For example: `pd-ssd`
