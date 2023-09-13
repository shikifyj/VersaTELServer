package linstor

import (
	"context"
	"fmt"

	"github.com/LINBIT/golinstor/client"
	log "github.com/sirupsen/logrus"
)

func GetNodeData(ctx context.Context, c *client.Client) []map[string]string {
	nodes, err := c.Nodes.GetAll(ctx)
	resources, err := c.Resources.GetResourceView(ctx)
	nodesInfo := []map[string]string{}
	if err != nil {
		log.Fatal(err)
	}
	for _, node := range nodes {
		resNum := 0
		for _, res := range resources {
			if res.NodeName == node.Name {
				resNum++
			}
		}
		sps, err := c.Nodes.GetStoragePools(ctx, node.Name)
		if err != nil {
			log.Fatal(err)
		}
		defaultInterface := node.NetInterfaces[0]
		addr := fmt.Sprintf("%s:%d (%s)", defaultInterface.Address, defaultInterface.SatellitePort, defaultInterface.SatelliteEncryptionType)

		nodeInfo := map[string]string{
			"name":           node.Name,
			"nodeType":       node.Type,
			"resourceNum":    fmt.Sprintf("%d", resNum),
			"storagePoolNum": fmt.Sprintf("%d", len(sps)),
			"addr":           addr,
			"status":         node.ConnectionStatus,
		}
		nodesInfo = append(nodesInfo, nodeInfo)
	}
	return nodesInfo
}

func DescribeNode(ctx context.Context, c *client.Client, nodename string) error {
	_, err := c.Nodes.Get(ctx, nodename)
	return err
}

func CreateNode(ctx context.Context, c *client.Client, name, ip, nodeType string) error {
	netInterfaces := []client.NetInterface{client.NetInterface{Name: "default", Address: ip, SatellitePort: 3366, SatelliteEncryptionType: "Plain"}}
	node := client.Node{Name: name, Type: nodeType, NetInterfaces: netInterfaces}
	err := c.Nodes.Create(ctx, node)
	return err
}

func DeleteNode(ctx context.Context, c *client.Client, nodename string) error {
	return c.Nodes.Delete(ctx, nodename)
}
