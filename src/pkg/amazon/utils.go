package amazon

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"log"
	"sort"
	"strconv"
)

type VolumeParams struct {
	Svc            *ec2.EC2
	AWSZone        string
	AWSVolumeType  string
	SnapshotId     *string
	K8SClusterName string
	K8SNamespace   string
	K8SVolumeName  string
	Size           *int64
}

func GetVolumeSize(value int64) string {
	return strconv.FormatInt(value, 10) + "Gi"
}

func GetLatestSnapshot(svc *ec2.EC2, SnapshotLabel string, SnapshotLabelValue string) (ec2.Snapshot, error) {
	input := &ec2.DescribeSnapshotsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:" + SnapshotLabel),
				Values: []*string{
					aws.String(SnapshotLabelValue),
				},
			},
		},
	}

	result, err := svc.DescribeSnapshots(input)
	if err != nil {
		return ec2.Snapshot{}, err
	}

	var completedSnapshots []*ec2.Snapshot
	for _, v := range result.Snapshots {
		if *v.State == "completed" {
			completedSnapshots = append(completedSnapshots, v)
		}
	}

	if len(completedSnapshots) == 0 {
		return ec2.Snapshot{}, errors.New(fmt.Sprintf("Not found any completed snapshots with tag:%v and value:%v", SnapshotLabel, SnapshotLabelValue))
	}

	// Sort by snapshots StartTime
	sort.Slice(completedSnapshots, func(i, j int) bool { return completedSnapshots[i].StartTime.After(*completedSnapshots[j].StartTime) })

	snapshot := *completedSnapshots[0]

	return snapshot, nil
}

func CreateEBSVolume(p VolumeParams) (ec2.Volume, error) {
	name := p.K8SClusterName + "-" + p.K8SVolumeName
	owned := "kubernetes.io/cluster/" + p.K8SClusterName
	clusterName := p.K8SClusterName
	namespace := p.K8SNamespace
	persistentVolumeName := p.K8SVolumeName

	input := &ec2.CreateVolumeInput{
		AvailabilityZone: aws.String(p.AWSZone),
		VolumeType:       aws.String(p.AWSVolumeType),
		Size:             p.Size,
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("volume"),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(name),
					},
					{
						Key:   aws.String("owned"),
						Value: aws.String(owned),
					},
					{
						Key:   aws.String("KubernetesCluster"),
						Value: aws.String(clusterName),
					},
					{
						Key:   aws.String("kubernetes.io/created-for/pvc/namespace"),
						Value: aws.String(namespace),
					},
					{
						Key:   aws.String("kubernetes.io/created-for/pv/name"),
						Value: aws.String(persistentVolumeName),
					},
				},
			},
		},
	}

	if p.SnapshotId != nil {
		input.SnapshotId = p.SnapshotId
	}

	result, err := p.Svc.CreateVolume(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			log.Println(aerr)
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			log.Println(err.Error())
		}
		return ec2.Volume{}, err
	}

	log.Printf("Volume created: %v", *result.VolumeId)

	return *result, nil
}
