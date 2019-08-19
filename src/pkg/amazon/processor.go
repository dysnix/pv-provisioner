package amazon

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
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
	const WorkerSleepTimeout = 30

	const AWSRegionEnvName = "AWS_REGION"
	const AWSZoneEnvName = "AWS_ZONE"
	const AWSVolumeTypeEnvName = "AWS_VOLUME_TYPE"
	const SnapshotLabelEnvName = "AWS_SNAPSHOT_LABEL"

	const K8SPersistentVolumeReclaimPolicyEnvName = "K8S_PERSISTENT_VOLUME_RECLAIM_POLICY"

	const K8SClusterNameEnvName = "K8S_CLUSTER_NAME"

	var volume ec2.Volume
	var volumeSize *int64

	var AWSRegion = env.GetEnvOrPanic(AWSRegionEnvName)
	var AWSZone = env.GetEnvOrPanic(AWSZoneEnvName)
	var AWSVolumeType = env.GetEnvOrPanic(AWSVolumeTypeEnvName)
	var SnapshotLabel = env.GetEnvOrPanic(SnapshotLabelEnvName)

	var K8SPersistentVolumeReclaimPolicy = env.GetEnvOrPanic(K8SPersistentVolumeReclaimPolicyEnvName)
	var K8SClusterName = env.GetEnvOrPanic(K8SClusterNameEnvName)

	if k8s.ReclaimPolicies[K8SPersistentVolumeReclaimPolicy] == "" {
		log.Panicf("Unsopported volume Reclaim Policy %v!", K8SPersistentVolumeReclaimPolicy)
	}

	// AWS connect
	sess, err := session.NewSession(&aws.Config{Region: aws.String(AWSRegion)})
	if err != nil {
		panic(err)
	}
	svc := ec2.New(sess)

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
		pvcs, err := clientset.CoreV1().PersistentVolumeClaims("").List(listOptions)

		if err != nil {
			log.Print(err)
			time.Sleep(WorkerSleepTimeout * time.Second)
			continue
		}

		for _, pvc := range pvcs.Items {
			if pvc.Status.Phase == v1.ClaimPending {
				log.Println("Found pending PVC: ", pvc.Name)

				pvName := pvc.Name
				K8SNamespace := pvc.Namespace

				_, err = clientset.CoreV1().PersistentVolumes().Get(pvc.Name, metav1.GetOptions{})
				if errors.IsNotFound(err) {
					log.Printf("PV %s in namespace %s not found, finding snapshot...\n", pvName, K8SNamespace)

					resource := pvc.Spec.Resources.Requests[v1.ResourceStorage]
					storageSize := resource.String()
					storageSizeInt, err := strconv.ParseInt(strings.Replace(storageSize, "Gi", "", -1), 10, 32)
					if err != nil {
						log.Panicf("Error parse integer of K8S storage size: %v", storageSize)
					}
					volumeSize = &storageSizeInt

					createAwsVolumeParams := VolumeParams{
						Svc:            svc,
						AWSZone:        AWSZone,
						AWSVolumeType:  AWSVolumeType,
						K8SClusterName: K8SClusterName,
						K8SNamespace:   K8SNamespace,
						K8SVolumeName:  pvName,
						Size:           volumeSize,
						SnapshotId:     nil,
					}

					SnapshotLabelValue := pvc.Labels[SnapshotLabel]

					if SnapshotLabelValue == "" {
						log.Printf("PVC %s in namespace %s does not have label %s\n", pvc.Name, pvc.Namespace, SnapshotLabel)
						continue
					}

					snapshot, err := GetLatestSnapshot(svc, SnapshotLabel, SnapshotLabelValue)
					if err != nil {
						log.Printf("Error get latest snapshot for PV %v: %v", pvName, err)

						volume, err = CreateEBSVolume(createAwsVolumeParams)
						if err != nil {
							log.Panicf("Error creating volume for PV %v: %v", pvName, err)
						}
					} else {
						log.Printf("Found snapshot: %v", *snapshot.SnapshotId)

						createAwsVolumeParams.SnapshotId = snapshot.SnapshotId
						volume, err = CreateEBSVolume(createAwsVolumeParams)
						if err != nil {
							log.Panicf("Error creating volume for PV %v: %v", pvName, err)
						}
					}

					createK8SPersistentVolumeParams := k8s.CreateVolumeRequest{
						Clientset:        clientset,
						PVName:           pvName,
						Region:           AWSRegion,
						Zone:             AWSZone,
						DiskId:           *volume.VolumeId,
						DiskSize:         GetVolumeSize(*volume.Size),
						K8SReclaimPolicy: K8SPersistentVolumeReclaimPolicy,
						Labels:           pvc.Labels,
					}

					pv, err := k8s.CreateAWSPersistentVolume(createK8SPersistentVolumeParams)
					if err != nil {
						log.Panicf("Error creating volume %v: %v", pvName, err)
					}

					log.Printf("PV %v created successfully", pv.Name)
				} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
					log.Printf("Error getting PV %s: %v\n", pvName, statusError.ErrStatus.Message)
				} else if err != nil {
					panic(err.Error())
				} else {
					log.Printf("PV %s already exist, please wait of bound", pvName)
				}
			}

		}

		time.Sleep(WorkerSleepTimeout * time.Second)
	}
}
