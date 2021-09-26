package linstor

import (
	"context"
	"github.com/LINBIT/golinstor/client"
	log "github.com/sirupsen/logrus"
	"strconv"
)

type DrivePool struct {
	Kind client.ProviderKind
	VG string
	LV string
}

func (dp *DrivePool)GetStoragePoolProps() map[string]string{
	switch dp.Kind {
	case "LVM":
		return map[string]string{"StorDriver/LvmVg":dp.VG}
	case "LVM_THIN":
		return map[string]string{"StorDriver/LvmVg":dp.VG,"StorDriver/ThinPool":dp.LV}
	}
	return nil
}


//func newDrivePool(kind, volume string) DrivePool {
//
//}


func GetSPData(ctx context.Context, c *client.Client) []map[string]string {
	resources, err := c.Resources.GetResourceView(ctx)
	spsInfo := []map[string]string{}
	if err != nil {
		log.Fatal(err)
	}
	sps,_ := c.Nodes.GetStoragePoolView(ctx)
	for _,sp := range sps{
		resNum := 0
		for _, res := range resources{
			for _,v := range res.Volumes{
				if v.StoragePoolName == sp.StoragePoolName{
					resNum ++
				}
			}
		}

		spInfo := map[string]string{
			"name" : sp.StoragePoolName,
			"node" : sp.NodeName,
			"resNum": strconv.Itoa(resNum),
			"driver" : sp.Props["StorDriver/LvmVg"],
			"poolName" : sp.Props["StorDriver/StorPoolName"],
			"freeCapacity" : FormatSize(sp.FreeCapacity),
			"totalCapacity" : FormatSize(sp.TotalCapacity),
			"supportsSnapshots" : strconv.FormatBool(sp.SupportsSnapshots),
		}
		if len(sp.Reports) == 0 {
			spInfo["state"] = "OK"
		}
		spsInfo = append(spsInfo, spInfo)
	}
	return spsInfo
}

func CreateSP(ctx context.Context, c *client.Client,spName,nodeName string, pool DrivePool) error{
	props := pool.GetStoragePoolProps()
	sp := client.StoragePool{StoragePoolName:spName,ProviderKind:pool.Kind, Props: props}
	return c.Nodes.CreateStoragePool(ctx,nodeName,sp)
}

func DeleteSP(ctx context.Context, c *client.Client, spName,nodeName string) error {
	return c.Nodes.DeleteStoragePool(ctx,nodeName,spName)
}
