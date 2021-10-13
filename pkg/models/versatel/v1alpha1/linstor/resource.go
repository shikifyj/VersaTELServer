package linstor

import (
	"context"
	"fmt"
	"github.com/LINBIT/golinstor/client"
	"strconv"
)


func GetResources(ctx context.Context, c *client.Client) []map[string]string{
	resources, _ := c.Resources.GetResourceView(ctx)
	resMap := map[string]map[string]string{}
	mirrorWay := map[string]int{}
	resArray := []map[string]string{}
	for _, res := range resources {
		resInfo := map[string]string{}
		_, exist := mirrorWay[res.Resource.Name]
		if !exist {
			mirrorWay[res.Resource.Name] = 1
		} else {
			mirrorWay[res.Resource.Name]++
		}


		for _, vol := range res.Volumes {
			if vol.State.DiskState == "Diskless" {
				mirrorWay[res.Resource.Name]--
				break
			} else {
				resInfo["name"] = res.Resource.Name
				resInfo["size"] = FormatSize(vol.AllocatedSizeKib)
				resInfo["deviceName"] = vol.DevicePath
				resInfo["status"] = vol.State.DiskState
				resInfo["mirrorWay"] = strconv.Itoa(mirrorWay[res.Resource.Name])
				resMap[res.Resource.Name] = resInfo
				break
			}
		}

	}
	for _,v := range resMap{
		resArray = append(resArray, v)
	}
	return resArray
}

func DescribeResource(ctx context.Context, c *client.Client, resname string) error {
	_, err := c.ResourceDefinitions.Get(ctx,resname)
	return err
}

func CreateResource(ctx context.Context, c *client.Client,resName string, spNames []string, Size string) error {
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
			return err
		}

		vd := client.VolumeDefinition{SizeKib: size}
		vdCreate := client.VolumeDefinitionCreate{VolumeDefinition: vd}
		err = c.ResourceDefinitions.CreateVolumeDefinition(ctx, resName, vdCreate)
		if err != nil {
			return err
		}
	} else {
		errInfo := fmt.Sprintf("resource definition : %v already exists", resName)
		Message := client.ApiCallError{client.ApiCallRc{RetCode: -1, Message: errInfo}}
		return Message
	}

	//创建resource
	var nodeName string
	sps, errSP := c.Nodes.GetStoragePoolView(ctx)
	if errSP != nil {
		return errSP
	}

	for _, spName := range spNames {
		fmt.Println(spName)
		for _, sp := range sps {
			if sp.StoragePoolName == spName {
				nodeName = sp.NodeName
			}
		}
		resProps := map[string]string{"StorPoolName": spName}
		res := client.Resource{Name: resName, NodeName: nodeName, Props: resProps}
		resCreate := client.ResourceCreate{Resource: res}
		err = c.Resources.Create(ctx, resCreate)
		if err != nil{
			return err
		}
	}
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
	err := c.ResourceDefinitions.Delete(ctx,resName)
	return err
}

func CreateDisklessResource(ctx context.Context, c *client.Client, resName, nodeName string) error {
	resProps := map[string]string{"StorPoolName": "DfltDisklessStorPool"}
	res := client.Resource{Name: resName, NodeName: nodeName, Props: resProps}
	resCreate := client.ResourceCreate{Resource: res}
	return c.Resources.Create(ctx, resCreate)
}