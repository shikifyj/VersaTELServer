package linstor

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/LINBIT/golinstor/client"
)

// func GetResources(ctx context.Context, c *client.Client) []map[string]string {
// 	resources, _ := c.Resources.GetResourceView(ctx)
// 	resMap := map[string]map[string]string{}
// 	mirrorWay := map[string]int{}
// 	resArray := []map[string]string{}
// 	for _, res := range resources {
// 		resInfo := map[string]string{}
// 		_, exist := mirrorWay[res.Resource.Name]
// 		if !exist {
// 			mirrorWay[res.Resource.Name] = 1
// 		} else {
// 			mirrorWay[res.Resource.Name]++
// 		}

// 		for _, vol := range res.Volumes {
// 			if vol.State.DiskState == "Diskless" {
// 				mirrorWay[res.Resource.Name]--
// 				break
// 			} else {
// 				resInfo["name"] = res.Resource.Name
// 				resInfo["size"] = FormatSize(vol.AllocatedSizeKib)
// 				resInfo["deviceName"] = vol.DevicePath
// 				resInfo["status"] = vol.State.DiskState
// 				resInfo["mirrorWay"] = strconv.Itoa(mirrorWay[res.Resource.Name])
// 				resMap[res.Resource.Name] = resInfo
// 				break
// 			}
// 		}

// 	}
// 	for _, v := range resMap {
// 		resArray = append(resArray, v)
// 	}
// 	return resArray
// }

func GetResources(ctx context.Context, c *client.Client) []map[string]string {
	resources, _ := c.Resources.GetResourceView(ctx)
	resMap := map[string]map[string]string{}
	mirrorWay := map[string]int{}
	resArray := []map[string]string{}
	for _, res := range resources {
		// fmt.Println(res.Resource.Flags)
		// fmt.Println(res.Resource.LayerObject.Drbd.Connections) // Connection
		// fmt.Println(res.Resource.LayerObject.Drbd.Connections) // 可能Unused是false，InUse是ture
		resInfo := map[string]string{}
		resName := res.Resource.Name
		_, exist := mirrorWay[res.Resource.Name]
		if !exist {
			mirrorWay[res.Resource.Name] = 1
		} else {
			mirrorWay[res.Resource.Name]++
		}
		for _, vol := range res.Volumes {
			resInfo["name"] = resName
			resInfo["size"] = FormatSize(vol.AllocatedSizeKib)
			resInfo["deviceName"] = vol.DevicePath
			resInfo["mirrorWay"] = strconv.Itoa(mirrorWay[res.Resource.Name])
			if res.CreateTimestamp != nil {
				resInfo["createTime"] = res.CreateTimestamp.Time.String()
			}
			if vol.State.DiskState == "Diskless" {
				mirrorWay[resName]--
				break
			} else if strings.Contains(vol.State.DiskState, "SyncTarget") {
				resInfo["status"] = "Synching"
				if _, exist := resMap[resName]; exist && resMap[resName]["State"] == "Unhealthy" {
					resMap[resName]["MirrorWay"] = strconv.Itoa(mirrorWay[res.Resource.Name])
				} else {
					resMap[resName] = resInfo
				}
			} else if vol.State.DiskState == "UpToDate" {
				resInfo["status"] = "Healthy"
				resInfo["mirrorWay"] = strconv.Itoa(mirrorWay[res.Resource.Name])
				if _, exist := resMap[resName]; !exist {
					resInfo["status"] = "Healthy"
					resMap[resName] = resInfo
				} else {
					resMap[resName]["mirrorWay"] = strconv.Itoa(mirrorWay[res.Resource.Name])
				}
			} else if vol.State.DiskState == "" {
				resInfo["status"] = "Unhealthy"
				resMap[resName] = resInfo
			}

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
		if *res.State.InUse {
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

		if *res.State.InUse {
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
