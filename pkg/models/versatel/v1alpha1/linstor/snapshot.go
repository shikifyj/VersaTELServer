package linstor

import (
	"context"
	"fmt"
	"github.com/LINBIT/golinstor/client"
	"strconv"
	"time"
)

func GetThinResources(ctx context.Context, c *client.Client) []map[string]string {
	DRBDResource, _ := c.Resources.GetResourceView(ctx)
	snapResource, _ := c.Resources.GetSnapshotView(ctx)
	var ResMap []map[string]string
	var resArray []map[string]string
	processedNames := make(map[string]bool)

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
			if _, processed := processedNames[res.Name]; !processed {
				resInfo := map[string]string{
					"name":    res.Name,
					"numbers": strconv.Itoa(count),
				}
				ResMap = append(ResMap, resInfo)
				processedNames[res.Name] = true
			}
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
	resInfoMap := make(map[string]map[string]interface{})

	for _, res := range snapshotResource {
		for _, snapshot := range res.Snapshots {
			if resInfo, ok := resInfoMap[snapshot.SnapshotName]; ok {
				resInfo["node"] = append(resInfo["node"].([]string), snapshot.NodeName)
			} else {
				if snapshot.CreateTimestamp != nil {
					resInfo := map[string]interface{}{
						"name":     snapshot.SnapshotName,
						"resource": res.ResourceName,
						"node":     []string{snapshot.NodeName},
						"state":    res.Flags[0],
						"time":     snapshot.CreateTimestamp.Time.Format("2006-01-02T15:04:05Z"),
					}

					resInfoMap[snapshot.SnapshotName] = resInfo
				}
			}
		}
	}
	for _, v := range resInfoMap {
		strMap := make(map[string]string)
		for key, value := range v {
			strMap[key] = fmt.Sprintf("%v", value)
		}
		ResMap = append(ResMap, strMap)
	}
	return ResMap
}

func CreateSnapshot(ctx context.Context, c *client.Client, resName string, snapName string) error {
	timestamp := time.Now().Format("20060102150405") // Format as YYYYMMDDHHMMSS
	snapNameWithTimestamp := fmt.Sprintf("%s_%s", snapName, timestamp)
	snapshot := client.Snapshot{
		Name:              snapNameWithTimestamp,
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
