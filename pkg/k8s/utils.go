package k8s

import (
	"context"
	"fmt"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	clientset "k8s.io/client-go/kubernetes"
	"log"
)

const K8SVolumeFSType = "ext4"
const K8SVolumeCreatedBy = "pv-provisioner"

var ReclaimPolicies = map[string]v1.PersistentVolumeReclaimPolicy{
	"Recycle": v1.PersistentVolumeReclaimRecycle,
	"Delete":  v1.PersistentVolumeReclaimDelete,
	"Retain":  v1.PersistentVolumeReclaimRetain,
}

type CreateVolumeRequest struct {
	Clientset        clientset.Interface
	PVName           string
	Namespace        string
	PDName           string
	Region           string
	Zone             string
	DiskId           string
	DiskSize         string
	K8SReclaimPolicy string
	Labels           map[string]string
}

type PersistentVolumeConfig struct {
	Name             string
	Namespace        string
	PVSource         v1.PersistentVolumeSource
	ReclaimPolicy    v1.PersistentVolumeReclaimPolicy
	NamePrefix       string
	Labels           labels.Set
	StorageClassName string
	NodeAffinity     *v1.VolumeNodeAffinity
	VolumeMode       *v1.PersistentVolumeMode
	StorageSize      string
}

func createPV(c clientset.Interface, pv *v1.PersistentVolume) (*v1.PersistentVolume, error) {
	pv, err := c.CoreV1().PersistentVolumes().Create(context.TODO(), pv, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("PV Create API error: %v", err)
	}
	return pv, nil
}

func MakePersistentVolume(pvConfig PersistentVolumeConfig) *v1.PersistentVolume {
	var claimRef *v1.ObjectReference
	// If the reclaimPolicy is not provided, assume Retain
	if pvConfig.ReclaimPolicy == "" {
		log.Printf("PV ReclaimPolicy unspecified, default: Retain")
		pvConfig.ReclaimPolicy = v1.PersistentVolumeReclaimRetain
	}
	claimRef = &v1.ObjectReference{
		Name:      pvConfig.Name,
		Namespace: pvConfig.Namespace,
	}
	return &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: pvConfig.NamePrefix,
			Labels:       pvConfig.Labels,
			Annotations: map[string]string{
				"kubernetes.io/createdby": K8SVolumeCreatedBy,
			},
			Finalizers: []string{"kubernetes.io/pv-protection"},
			Name:       pvConfig.Name,
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: pvConfig.ReclaimPolicy,
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): resource.MustParse(pvConfig.StorageSize),
			},
			PersistentVolumeSource: pvConfig.PVSource,
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},
			ClaimRef:         claimRef,
			StorageClassName: pvConfig.StorageClassName,
			NodeAffinity:     pvConfig.NodeAffinity,
			VolumeMode:       pvConfig.VolumeMode,
		},
	}
}

func CreateGCPPersistentVolume(r CreateVolumeRequest) (*v1.PersistentVolume, error) {
	pvConfig := PersistentVolumeConfig{
		Name:      r.PVName,
		Namespace: r.Namespace,
		PVSource: v1.PersistentVolumeSource{
			GCEPersistentDisk: &v1.GCEPersistentDiskVolumeSource{
				PDName: r.PDName,
				FSType: K8SVolumeFSType,
			},
		},
		StorageSize:   r.DiskSize,
		ReclaimPolicy: ReclaimPolicies[r.K8SReclaimPolicy],
		Labels:        r.Labels,
	}
	pv := MakePersistentVolume(pvConfig)
	pv, err := createPV(r.Clientset, pv)
	if err != nil {
		return nil, fmt.Errorf("PV Create API error: %v", err)
	}
	return pv, nil
}

func getVolumeId(zone string, volumeId string) string {
	return "aws://" + zone + "/" + volumeId
}

func CreateAWSPersistentVolume(r CreateVolumeRequest) (*v1.PersistentVolume, error) {
	r.Labels["failure-domain.beta.kubernetes.io/region"] = r.Region
	r.Labels["failure-domain.beta.kubernetes.io/zone"] = r.Zone

	pvConfig := PersistentVolumeConfig{
		Name:      r.PVName,
		Namespace: r.Namespace,
		PVSource: v1.PersistentVolumeSource{
			AWSElasticBlockStore: &v1.AWSElasticBlockStoreVolumeSource{
				VolumeID: getVolumeId(r.Zone, r.DiskId),
				FSType:   K8SVolumeFSType,
			},
		},
		StorageSize:   r.DiskSize,
		ReclaimPolicy: ReclaimPolicies[r.K8SReclaimPolicy],
		Labels:        r.Labels,
	}
	pv := MakePersistentVolume(pvConfig)
	pv, err := createPV(r.Clientset, pv)
	if err != nil {
		return nil, fmt.Errorf("PV Create API error: %v", err)
	}
	return pv, nil
}
