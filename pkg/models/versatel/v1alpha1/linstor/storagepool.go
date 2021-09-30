package linstor

import (
	"context"
	"github.com/LINBIT/golinstor/client"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
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
			"driver" : string(sp.ProviderKind),
			"poolName" : sp.Props["StorDriver/StorPoolName"],
			"freeCapacity" : FormatSize(sp.FreeCapacity),
			"totalCapacity" : FormatSize(sp.TotalCapacity),
			"supportsSnapshots" : strconv.FormatBool(sp.SupportsSnapshots),
		}
		if len(sp.Reports) == 0 {
			spInfo["status"] = "OK"
		}
		spsInfo = append(spsInfo, spInfo)
	}
	return spsInfo
}

func CreateSP(ctx context.Context, c *client.Client,spName,nodeName,kind,volume string) error{
	pool := DrivePool{}
	if (kind == "LVM" || kind == "lvm"){
		pool.Kind = client.LVM
		pool.VG = volume
	} else if (kind == "LVM THIN" || kind == "lvm thin") {
		pool.Kind = client.LVM_THIN
		volSlice := strings.Split(volume,"/")
		pool.VG = volSlice[0]
		pool.LV = volSlice[1]
	}
	props := pool.GetStoragePoolProps()
	sp := client.StoragePool{StoragePoolName:spName,ProviderKind:pool.Kind, Props: props}
	return c.Nodes.CreateStoragePool(ctx,nodeName,sp)
}

func DeleteSP(ctx context.Context, c *client.Client, spName,nodeName string) error {
	return c.Nodes.DeleteStoragePool(ctx,nodeName,spName)
}
