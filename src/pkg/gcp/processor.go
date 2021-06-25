package gcp

import (
	"context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"pkg/env"
	"pkg/k8s"
	"strconv"
	"strings"
	"time"
)

func RunProcessor() {
	const WorkerSleepTimeout = 10

	const ProjectEnvName = "GCP_PROJECT"
	const ZonesEnvName = "GCP_ZONES"
	const SnapshotLabelEnvName = "GCP_SNAPSHOT_LABEL"
	const GCPDiskTypeEnvName = "GCP_DISK_TYPE"

	const K8SPersistentVolumeReclaimPolicyEnvName = "K8S_PERSISTENT_VOLUME_RECLAIM_POLICY"

	var Project = env.GetEnvOrPanic(ProjectEnvName)
	var GCPZones = env.GetEnvOrPanic(ZonesEnvName)
	var GCPSnapshotLabel = env.GetEnvOrPanic(SnapshotLabelEnvName)
	var GCPDiskType = env.GetEnvOrPanic(GCPDiskTypeEnvName)

	var K8SPersistentVolumeReclaimPolicy = env.GetEnvOrPanic(K8SPersistentVolumeReclaimPolicyEnvName)

	if k8s.ReclaimPolicies[K8SPersistentVolumeReclaimPolicy] == "" {
		log.Panicf("Unsopported volume ReclaimPolicy %v!", K8SPersistentVolumeReclaimPolicy)
	}

	// GCP connect
	ctx := context.Background()
	client, err := google.DefaultClient(ctx, compute.ComputeScope)
	if err != nil {
		log.Fatal(err)
	}

	svc, err := compute.New(client)
	if err != nil {
		log.Fatal(err)
	}

	// Kubernetes connect
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	for {
		listOptions := metav1.ListOptions{}
		pvcs, err := clientset.CoreV1().PersistentVolumeClaims("").List(context.TODO(), listOptions)

		if err != nil {
			log.Print(err)
			time.Sleep(WorkerSleepTimeout * time.Second)
			continue
		}

		for _, pvc := range pvcs.Items {
			if pvc.Status.Phase == v1.ClaimPending {
				log.Println("Found pending PVC: ", pvc.Name)

				if *pvc.Spec.StorageClassName != "" {
					log.Print("Pass PVC with StorageClass...")
					continue
				}

				_, err = clientset.CoreV1().PersistentVolumes().Get(context.TODO(), pvc.Name, metav1.GetOptions{})

				if errors.IsNotFound(err) {
					log.Printf("PV for PVC %s in namespace %s not found, finding snapshot...\n", pvc.Name, pvc.Namespace)

					resource := pvc.Spec.Resources.Requests[v1.ResourceStorage]
					storageSize := resource.String()
					storageSizeInt, err := strconv.ParseInt(strings.Replace(storageSize, "Gi", "", -1), 10, 32)
					if err != nil {
						log.Panicf("Error parse integer of K8S storage size: %v", storageSize)
					}

					createGCPVolumeParams := VolumeParams{
						Svc:            svc,
						Zone:           getRandomZone(GCPZones),
						Project:        Project,
						Name:           pvc.Name,
						Size:           storageSizeInt,
						Type:           GCPDiskType,
						SourceSnapshot: "",
					}

					GCPSnapshotLabelValue := pvc.Labels[GCPSnapshotLabel]

					if GCPSnapshotLabelValue == "" {
						log.Printf("PVC %s in namespace %s does not have label %s\n", pvc.Name, pvc.Namespace, GCPSnapshotLabel)
						continue
					}

					snapshot, err := GetLatestSnapshot(ctx, svc, Project, GCPSnapshotLabel, GCPSnapshotLabelValue)
					if err != nil {
						log.Printf("Error get latest snapshot for PV %v: %v", pvc.Name, err)
					} else {
						log.Printf("Found snapshot: %v", snapshot.Name)
						createGCPVolumeParams.SourceSnapshot = snapshot.SelfLink
					}

					disk, err := CreateGCPDisk(createGCPVolumeParams)
					if err != nil {
						log.Panicf("Error creating disk for PV %v: %v", pvc.Name, err)
					}

					createK8SPersistentVolumeParams := k8s.CreateVolumeRequest{
						Clientset:        clientset,
						PVName:           pvc.Name,
						Namespace:        pvc.Namespace,
						PDName:           disk.Name,
						DiskSize:         GetDiskSize(disk.SizeGb),
						K8SReclaimPolicy: K8SPersistentVolumeReclaimPolicy,
						Labels:           pvc.Labels,
					}

					pv, err := k8s.CreateGCPPersistentVolume(createK8SPersistentVolumeParams)
					if err != nil {
						log.Panicf("Error creating volume %v: %v", pvc.Name, err)
					}

					log.Printf("PV %v created successfully", pv.Name)
				} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
					log.Printf("Error getting PV %s: %v\n", pvc.Name, statusError.ErrStatus.Message)
				} else if err != nil {
					panic(err.Error())
				} else {
					log.Printf("PV %s already exist, please wait of bound", pvc.Name)
				}
			}

		}

		time.Sleep(WorkerSleepTimeout * time.Second)
	}
}
