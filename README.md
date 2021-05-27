# pv-provisioner

Kubernetes Persistent Volume Provisioner from pre-created snapshots.

## How its works?

* You create PersistentVolumeClaim with empty Storage class (for disable [Dynamic Provisioning](https://kubernetes.io/docs/concepts/storage/dynamic-provisioning/#enabling-dynamic-provisioning)
* _pv-provisioner_ detect PVC and track tag `app` value
* _pv-provisioner_ search most actual snapshot with label `app` value equal `app` tag of k8s PVC
* _pv-provisioner_ create new Disk/Volume
* _pv-provisioner_ create new PersistentVolume with required params
* PVC use created PV

## Usage

    pv-provisioner -cloud <cloud>

## Supported cloud platforms:

* "gcp" - Google Cloud Platform

### GCP Environment variables

* **GCP_PROJECT** - GCP project ID
* **GCP_ZONE** - GCP Disks availability zone
* **GCP_SNAPSHOT_LABEL** - tag name for find Snapshot
* **GCP_DISK_TYPE_URL** - URL of the disk type resource describing which disk type to use to create the disk. Provide this when creating the disk. For example: projects/project/zones/zone/diskTypes/pd-standard or pd-ssd

## TODO

* Use [Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity) for GCP