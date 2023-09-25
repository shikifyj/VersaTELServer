package linstor

import (
	"context"
	"fmt"
	"github.com/LINBIT/golinstor/client"
	"strconv"
)

func GetThinResources(ctx context.Context, c *client.Client) []map[string]string {
	DRBDResource, _ := c.Resources.GetResourceView(ctx)
	snapResource, _ := c.Resources.GetSnapshotView(ctx)
	var ResMap []map[string]string
	var resArray []map[string]string

	for _, res := range DRBDResource {
		for _, vol := range res.Volumes {
			if string(vol.ProviderKind) != "LVM_THIN" {
				continue
			}
			count := 0
			for _, snap := range snapResource {
				if snap.ResourceName == res.Name {
					count++
				}
			}
			resInfo := map[string]string{
				"name":    res.Name,
				"numbers": strconv.Itoa(count),
			}
			ResMap = append(ResMap, resInfo)

		}
	}
	for _, v := range ResMap {
		resArray = append(resArray, v)
	}
	return resArray
}

func GetSnapshot(ctx context.Context, c *client.Client) []map[string]string {
	snapshotResource, _ := c.Resources.GetSnapshotView(ctx)
	var ResMap []map[string]string
	var resArray []map[string]string

	for _, res := range snapshotResource {
		for _, snapshot := range res.Snapshots {
			resInfo := map[string]string{
				"name":     snapshot.SnapshotName,
				"resource": res.ResourceName,
				"time":     snapshot.CreateTimestamp.Time.String(),
				"state":    res.Flags[0],
			}
			ResMap = append(ResMap, resInfo)
		}
	}
	for _, v := range ResMap {
		resArray = append(resArray, v)
	}
	return resArray
}

func CreateSnapshot(ctx context.Context, c *client.Client, resName string, snapName string) error {
	snapshot := client.Snapshot{
		Name:              snapName,
		ResourceName:      resName,
		Nodes:             nil,
		Props:             nil,
		Flags:             nil,
		VolumeDefinitions: nil,
		Uuid:              "",
		Snapshots:         nil,
	}
	err := c.Resources.CreateSnapshot(ctx, snapshot)
	return err
}

func DeleteSnapshot(ctx context.Context, c *client.Client, resName string, snapName string) error {
	err := c.Resources.DeleteSnapshot(ctx, resName, snapName)
	return err
}

func RollbackSnapshot(ctx context.Context, c *client.Client, resName string, snapName string) error {
	err := c.Resources.RollbackSnapshot(ctx, resName, snapName)
	return err
}

func RecoverSnapshot(ctx context.Context, c *client.Client, resName string, snapName string, newResName string, Nodes []string) error {
	SnapshotRestore := client.SnapshotRestore{
		ToResource: newResName,
		Nodes:      Nodes,
	}
	if _, err := c.ResourceDefinitions.Get(ctx, newResName); err != nil {
		rd := client.ResourceDefinition{Name: newResName}
		rdCreate := client.ResourceDefinitionCreate{ResourceDefinition: rd}
		err = c.ResourceDefinitions.Create(ctx, rdCreate)
		if err != nil {
			c.ResourceDefinitions.Delete(ctx, newResName)
			return err
		}
	} else {
		errInfo := fmt.Sprintf("resource definition : %v already exists", resName)
		Message := client.ApiCallError{client.ApiCallRc{RetCode: -1, Message: errInfo}}
		return Message
	}
	err := c.Resources.RestoreVolumeDefinitionSnapshot(ctx, resName, snapName, SnapshotRestore)
	if err == nil {
		err := c.Resources.RestoreSnapshot(ctx, resName, snapName, SnapshotRestore)
		fmt.Println(err)
		return err
	}
	return err
}
