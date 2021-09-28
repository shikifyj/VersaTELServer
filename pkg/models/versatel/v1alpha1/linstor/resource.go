package linstor

import (
	"context"
	"fmt"
	"github.com/LINBIT/golinstor/client"
	"strconv"
)


func GetResources(ctx context.Context, c *client.Client) []map[string]string{
	resources, _ := c.Resources.GetResourceView(ctx)
	resourcesInfo := []map[string]string{}
	for _, res := range resources{
		name := res.Resource.Name
		mirrorway := strconv.Itoa(len(res.Volumes))
		for _, vol := range res.Volumes{
			resInfo := map[string]string{}
			resInfo["Name"] = name
			resInfo["MirrorWay"] = mirrorway
			resInfo["Size"] = FormatSize(vol.AllocatedSizeKib)
			resInfo["DeviceName"] = vol.DevicePath
			resInfo["State"] = vol.State.DiskState
			resourcesInfo = append(resourcesInfo, resInfo)
		}
	}
	return resourcesInfo
}

func CreateResource(ctx context.Context, c *client.Client,resName,nodeName,spName string,SizeKib uint64,) error {
	// VolNr 应该可以在Resource的Props设置
	var err error

	//创建 rd
	if _, err := c.ResourceDefinitions.Get(ctx,resName); err != nil{
		rd := client.ResourceDefinition{Name: resName}
		rdCreate := client.ResourceDefinitionCreate{ResourceDefinition: rd}
		err = c.ResourceDefinitions.Create(ctx,rdCreate)
		if err != nil{
			return err
		}

		vd := client.VolumeDefinition{SizeKib: SizeKib}
		vdCreate := client.VolumeDefinitionCreate{VolumeDefinition: vd}
		err = c.ResourceDefinitions.CreateVolumeDefinition(ctx,resName, vdCreate)
		if err != nil{
			return err
		}
	} else {
		errInfo := fmt.Sprintf("resource definition : %v already exists",resName)
		Message := client.ApiCallError{client.ApiCallRc{RetCode: -1, Message: errInfo}}
		return Message
	}

	//创建resource
	resProps := map[string]string{"StorPoolName":spName}
	res := client.Resource{Name:resName,NodeName: nodeName,Props: resProps}
	resCreate := client.ResourceCreate{Resource: res}
	err = c.Resources.Create(ctx,resCreate)
	return err
}

func DeleteResource(ctx context.Context, c *client.Client, resName,nodeName string) error {
	err := c.Resources.Delete(ctx,resName,nodeName)
	if err != nil{
		return err
	}
	resources, _ := c.Resources.GetAll(ctx,resName)
	if len(resources) == 0 {
		err = c.ResourceDefinitions.Delete(ctx,resName)
	}
	return err
}