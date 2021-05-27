package gcp

import (
	"context"
	"google.golang.org/api/compute/v1"
	"log"
	"sort"
	"strconv"
	"time"
)

type VolumeParams struct {
	Svc            *compute.Service
	Name           string
	Zone           string
	Project        string
	SourceSnapshot string
	Type           string
	Size           int64
}

func GetDiskSize(value int64) string {
	return strconv.FormatInt(value, 10) + "Gi"
}

func GetLatestSnapshot(ctx context.Context, svc *compute.Service, project string, labelName string, labelValue string) (compute.Snapshot, error) {
	req := svc.Snapshots.List(project).Filter("labels." + labelName + "=" + labelValue)

	snapshots := []compute.Snapshot{}

	if err := req.Pages(ctx, func(page *compute.SnapshotList) error {
		for _, snapshot := range page.Items {
			if snapshot.Status == "READY" {
				snapshots = append(snapshots, *snapshot)
			}
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}

	layout := "2018-12-19T02:36:19.635-08:00"
	sort.Slice(snapshots, func(i, j int) bool {
		ti, _ := time.Parse(layout, snapshots[i].CreationTimestamp)
		tj, _ := time.Parse(layout, snapshots[j].CreationTimestamp)
		return ti.After(tj)
	})

	snapshot := snapshots[0]

	return snapshot, nil
}

func CreateGCPDisk(p VolumeParams) (compute.Disk, error) {
	disk := compute.Disk{
		Name:   p.Name,
		SizeGb: p.Size,
		Type:   p.Type,
	}

	if p.SourceSnapshot != "" {
		disk.SourceSnapshot = p.SourceSnapshot
	}

	result, err := p.Svc.Disks.Insert(p.Project, p.Zone, &disk).Do()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Volume created: %v", result.Name)

	return disk, nil
}
