package linstor

import (
	"context"
	"fmt"
	"github.com/LINBIT/golinstor/client"
	"sort"
	"strconv"
	"strings"
)

func GetResources(ctx context.Context, c *client.Client) []map[string]string {
	resources, _ := c.Resources.GetResourceView(ctx)
	resMap := make(map[string]map[string]string)
	mirrorWay := make(map[string]int)
	resArray := []map[string]string{}

	for _, res := range resources {
		resName := res.Resource.Name
		mirrorWay[resName]++

		for _, vol := range res.Volumes {
			resInfo := map[string]string{
				"name":         resName,
				"size":         FormatSize(vol.AllocatedSizeKib),
				"deviceName":   vol.DevicePath,
				"mirrorWay":    strconv.Itoa(mirrorWay[res.Resource.Name]),
				"assignedNode": "",
			}

			if res.CreateTimestamp != nil {
				resInfo["createTime"] = res.CreateTimestamp.Time.String()
			}

			connFail := map[string][]string{}
			for k, v := range res.LayerObject.Drbd.Connections {
				if !v.Connected {
					connFail[v.Message] = append(connFail[v.Message], k)
				}
			}

			volState := res.Volumes[0].State.DiskState
			if volState == "Diskless" {
				mirrorWay[resName]--
				break
			} else if volState == "UpToDate" && len(connFail) != 0 {
				resInfo["status"] = "Unhealthy"
				resMap[resName] = resInfo
			} else if volState == "UpToDate" && len(connFail) == 0 {
				resInfo["status"] = "Healthy"
				resMap[resName] = resInfo
			} else if volState == "" {
				resInfo["status"] = "Unhealthy"
				resMap[resName] = resInfo
			} else if strings.Contains(vol.State.DiskState, "SyncTarget") {
				resInfo["status"] = "Synching"
				if _, exist := resMap[resName]; exist && resMap[resName]["State"] == "Unhealthy" {
					resMap[resName]["MirrorWay"] = strconv.Itoa(mirrorWay[res.Resource.Name])
				} else {
					resMap[resName] = resInfo
				}
			}
			//switch volState {
			//case "Diskless":
			//	mirrorWay[resName]--
			//case "UpToDate":
			//	resInfo["status"] = "Healthy"
			//	resMap[resName] = resInfo
			//case "":
			//	resInfo["status"] = "Unhealthy"
			//	resMap[resName] = resInfo
			//default:
			//	resInfo["status"] = "Synching"
			//	if _, exist := resMap[resName]; exist && resMap[resName]["status"] == "Unhealthy" {
			//		resMap[resName]["mirrorWay"] = strconv.Itoa(mirrorWay[resName])
			//	} else {
			//		resMap[resName] = resInfo
			//	}
			//}

		}
	}

	for _, v := range resMap {
		resArray = append(resArray, v)
	}

	sort.SliceStable(resArray, func(i, j int) bool {
		return resArray[i]["createTime"] > resArray[j]["createTime"]
	})

	return resArray
}

func GetassignedNode(ctx context.Context, c *client.Client) []map[string]string {
	resArray := []map[string]string{}
	resources, _ := c.Resources.GetResourceView(ctx)
	resMap := make(map[string]map[string]string)
	for _, res := range resources {
		resName := res.Resource.Name
		for _, vol := range res.Volumes {
			resInfo := map[string]string{
				"name": resName,
			}
			if vol.State.DiskState == "Diskless" {
				resInfo["assignedNode"] = res.Resource.NodeName
				resMap[resName] = resInfo
			}
		}
	}
	for _, v := range resMap {
		resArray = append(resArray, v)
	}
	return resArray
}

func GetResourcesDiskful(ctx context.Context, c *client.Client) []map[string]string {
	resources, _ := c.Resources.GetResourceView(ctx)
	resMap := []map[string]string{}
	resMap = append(resMap)

	for _, res := range resources {
		mapFlag := make(map[string]struct{}, len(res.Flags))
		for _, v := range res.Flags {
			mapFlag[v] = struct{}{}
		}

		resInfo := map[string]string{}

		if res.State != nil {
			if *res.State.InUse {
				resInfo["usage"] = "InUse"
			} else {
				resInfo["usage"] = "Unused"
			}
		} else {
			resInfo["usage"] = "Unknown"
		}

		connFail := map[string][]string{}
		for k, v := range res.LayerObject.Drbd.Connections {
			if !v.Connected {
				connFail[v.Message] = append(connFail[v.Message], k)
			}
		}
		if len(connFail) != 0 {
			connFailStr := make([]string, 0)
			for k, v := range connFail {
				connFailStr = append(connFailStr, k+"("+strings.Join(v, ",")+")")
			}
			resInfo["conn"] = strings.Join(connFailStr, ",")
		} else {
			resInfo["conn"] = "OK"
		}

		resInfo["status"] = "Unknow"
		for _, vol := range res.Volumes {
			if vol.State.DiskState != "Diskless" {
				if _, ok := mapFlag["DELETE"]; ok {
					resInfo["status"] = "DELETING"
				} else if _, ok := mapFlag["INACTIVE"]; ok {
					resInfo["status"] = "INACTIVE"
				} else if vol.State.DiskState != "" {
					resInfo["status"] = vol.State.DiskState
				}
				resInfo["name"] = res.Resource.Name
				resInfo["node"] = res.Resource.NodeName
				resInfo["storagepool"] = vol.StoragePoolName
				resMap = append(resMap, resInfo)
				break
			}
		}

	}
	return resMap
}

func GetResourceDiskless(ctx context.Context, c *client.Client) []map[string]string {
	resources, _ := c.Resources.GetResourceView(ctx)
	resMap := []map[string]string{}

	for _, res := range resources {
		resInfo := map[string]string{}

		if res.State != nil && *res.State.InUse {
			resInfo["usage"] = "InUse"
		} else {
			resInfo["usage"] = "Unused"
		}

		connFail := map[string][]string{}
		for k, v := range res.LayerObject.Drbd.Connections {
			if !v.Connected {
				connFail[v.Message] = append(connFail[v.Message], k)
			}
		}
		if len(connFail) != 0 {
			connFailStr := make([]string, 0)
			for k, v := range connFail {
				connFailStr = append(connFailStr, k+"("+strings.Join(v, ",")+")")
			}
			resInfo["conn"] = strings.Join(connFailStr, ",")
		} else {
			resInfo["conn"] = "OK"
		}

		for _, vol := range res.Volumes {
			if vol.State.DiskState == "Diskless" {
				resInfo["status"] = "Diskless"
				resInfo["name"] = res.Resource.Name
				resInfo["node"] = res.Resource.NodeName
				resInfo["storagepool"] = vol.StoragePoolName
				resMap = append(resMap, resInfo)
				break
			}
		}

	}
	return resMap
}

func DescribeResource(ctx context.Context, c *client.Client, resname string) error {
	_, err := c.ResourceDefinitions.Get(ctx, resname)
	return err
}

func CreateResource(ctx context.Context, c *client.Client, resName string, sps [][]string, Size string) error {
	// VolNr 应该可以在Resource的Props设置
	var err error
	size, errSize := ParseSize(Size)
	if errSize != nil {
		Message := client.ApiCallError{client.ApiCallRc{RetCode: -2, Message: "Error Size"}}
		return Message
	}

	//创建 rd
	if _, err := c.ResourceDefinitions.Get(ctx, resName); err != nil {
		rd := client.ResourceDefinition{Name: resName}
		rdCreate := client.ResourceDefinitionCreate{ResourceDefinition: rd}
		err = c.ResourceDefinitions.Create(ctx, rdCreate)
		if err != nil {
			c.ResourceDefinitions.Delete(ctx, resName)
			return err
		}

		vd := client.VolumeDefinition{SizeKib: size}
		vdCreate := client.VolumeDefinitionCreate{VolumeDefinition: vd}
		err = c.ResourceDefinitions.CreateVolumeDefinition(ctx, resName, vdCreate)
		if err != nil {
			c.ResourceDefinitions.Delete(ctx, resName)
			return err
		}
	} else {
		errInfo := fmt.Sprintf("resource definition : %v already exists", resName)
		Message := client.ApiCallError{client.ApiCallRc{RetCode: -1, Message: errInfo}}
		return Message
	}

	//创建resource
	var nodeName string
	var spName string
	for _, spAndNode := range sps {
		spName = spAndNode[0]
		nodeName = spAndNode[1]
		resProps := map[string]string{"StorPoolName": spName}
		res := client.Resource{Name: resName, NodeName: nodeName, Props: resProps}
		resCreate := client.ResourceCreate{Resource: res}
		err = c.Resources.Create(ctx, resCreate)
		if err != nil {
			c.ResourceDefinitions.Delete(ctx, resName)
			return err
		}
	}

	// sps, _ := c.Nodes.GetStoragePoolView(ctx)

	// for _, spName := range spNames {
	// 	for _, sp := range sps {
	// 		if sp.StoragePoolName == spName {
	// 			nodeName = sp.NodeName
	// 		}
	// 	}
	// 	resProps := map[string]string{"StorPoolName": spName}
	// 	res := client.Resource{Name: resName, NodeName: nodeName, Props: resProps}
	// 	resCreate := client.ResourceCreate{Resource: res}
	// 	err = c.Resources.Create(ctx, resCreate)
	// 	if err != nil {
	// 		c.ResourceDefinitions.Delete(ctx, resName)
	// 		return err
	// 	}
	// }
	return nil
}

func DeleteResource(ctx context.Context, c *client.Client, resName string) error {
	//err := c.Resources.Delete(ctx,resName,nodeName)
	//if err != nil{
	//	return err
	//}
	//resources, _ := c.Resources.GetAll(ctx,resName)
	//if len(resources) == 0 {
	//	err = c.ResourceDefinitions.Delete(ctx,resName)
	//}
	err := c.ResourceDefinitions.Delete(ctx, resName)
	return err
}

func CreateDisklessResource(ctx context.Context, c *client.Client, resName, nodeName string) error {
	resProps := map[string]string{"StorPoolName": "DfltDisklessStorPool"}
	res := client.Resource{Name: resName, NodeName: nodeName, Props: resProps}
	resCreate := client.ResourceCreate{Resource: res}
	return c.Resources.Create(ctx, resCreate)
}

func UpdateDiskfulResource(ctx context.Context, c *client.Client, resName string, nodeName []string, storagePoolName []string,
	targetReplicas int, currentReplicas int) error {
	// 副本数量差
	delta := currentReplicas - targetReplicas

	if delta > 0 {
		for i := range nodeName {
			nName := nodeName[i]
			spName := storagePoolName[i]

			resProps := map[string]string{"StorPoolName": spName}
			newRes := client.Resource{Name: resName, NodeName: nName, Props: resProps}
			resCreate := client.ResourceCreate{Resource: newRes}
			err := c.Resources.Create(ctx, resCreate)
			if err != nil {
				return err
			}
		}
	} else if delta < 0 {
		for _, nName := range nodeName {
			resources, err := c.Resources.GetAll(ctx, resName)
			if err != nil {
				return err
			}
			nodeNames := []string{}
			for _, res := range resources {
				fmt.Println(res.NodeName)
				nodeNames = append(nodeNames, res.NodeName)
			}
			if !contains(nodeNames, nName) {
				errInfo := fmt.Sprintf("在节点 %s 下不存在副本 %s ,请重新选择节点调整副本", nName, resName)
				Message := client.ApiCallError{client.ApiCallRc{RetCode: -1, Message: errInfo}}
				return Message
			}
			err = c.Resources.Delete(ctx, resName, nName)
			if err != nil {
				return err
			}
			err = c.Resources.Delete(ctx, resName, nName)
			if err != nil {
				return err
			}
		}
	} else {
		return nil
	}

	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
