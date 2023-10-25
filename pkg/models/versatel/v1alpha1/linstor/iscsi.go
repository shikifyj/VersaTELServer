package linstor

import (
	"context"
	"fmt"
	"github.com/LINBIT/golinstor/client"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

func GetIPAndConnect(port int) (*ssh.Client, error) {
	ipFile := "/etc/linstorip/linstorip"

	ipBytes, err := ioutil.ReadFile(ipFile)
	if err != nil {
		log.Fatalf("unable to read IP file: %v", err)
		return nil, err
	}

	ip := strings.TrimSpace(string(ipBytes))

	sshClient, err := SSHConnect(ip, port)
	if err != nil {
		log.Fatalf("Failed to connect to remote host: %v", err)
		return nil, err
	}

	return sshClient, nil
}

func CreateYamlFile(filename string, data string) {
	filePath := "/etc/linstorip/" + filename
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("Failed to create file: %s", err)
	}
	defer file.Close()

	_, err = file.WriteString(data)
	if err != nil {
		log.Fatalf("Failed to write to file: %s", err)
	}

	err = file.Sync()
	if err != nil {
		log.Fatalf("Failed to save file: %s", err)
	}
}

func CreatePortBlockOn(vips []string, tgn string) error {
	sc, _ := GetIPAndConnect(22)
	for vip := range vips {
		cmd := fmt.Sprintf("crm conf primitive vip_prtblk_on%s_1 portblock "+
			"prams ip=%s portno=3260 protocol=tcp action=block "+
			"op start timeout=20 interval=0 "+
			"op stop timeout=20 interval=0 "+
			"op monitor timeout=20 interval=20", tgn, strconv.Itoa(vip))
		out, err := SshCmd(sc, cmd)
		if err != nil {
			errInfo := fmt.Sprintf(strings.Replace(strings.TrimSpace(out), "\n", "", -1))
			Message := client.ApiCallError{client.ApiCallRc{Message: errInfo}}
			return Message
		}
	}

	return nil
}

func CreateVip(vips []string, tgn string) error {
	sc, _ := GetIPAndConnect(22)
	for vip := range vips {
		cmd := fmt.Sprintf("crm conf primitive vip%s_1 IPaddr2 "+
			"params ip=%s cidr_netmask=24 "+
			"op monitor interval=10 timeout=20", tgn, strconv.Itoa(vip))
		out, err := SshCmd(sc, cmd)
		if err != nil {
			errInfo := fmt.Sprintf(strings.Replace(strings.TrimSpace(out), "\n", "", -1))
			Message := client.ApiCallError{client.ApiCallRc{Message: errInfo}}
			return Message
		}
	}

	return nil
}

func CreateTarget(vips []string, tgn string, iqn string) error {
	sc, _ := GetIPAndConnect(22)
	var cmd string
	if len(vips) == 1 {
		cmd = fmt.Sprintf("crm conf primitive target%s iSCSITarget "+
			"params iqn=%s implementation=lio-t portals=\"%s:3260\" "+
			"op start timeout=50 interval=0 "+
			"op stop timeout=40 interval=0 "+
			"op monitor interval=15 timeout=40", tgn, iqn, vips[0])
	} else {
		cmd = fmt.Sprintf("crm conf primitive target%s iSCSITarget "+
			"params iqn=%s implementation=lio-t portals=\"%s:3260 %s:3260\" "+
			"op start timeout=50 interval=0 "+
			"op stop timeout=40 interval=0 "+
			"op monitor interval=15 timeout=40", tgn, iqn, vips[0], vips[1])
	}
	out, err := SshCmd(sc, cmd)
	if err != nil {
		errInfo := fmt.Sprintf(strings.Replace(strings.TrimSpace(out), "\n", "", -1))
		Message := client.ApiCallError{client.ApiCallRc{Message: errInfo}}
		return Message
	}

	return nil
}

func CreateResourceGroup(vips []string, tgn string, nodeLess []string) error {
	sc, _ := GetIPAndConnect(22)
	var cmd string
	if len(vips) == 1 {
		cmd = fmt.Sprintf("crm conf group gvip%s vip_prtblk_on%s_1 vip%s_1 target%s "+
			"meta target-role=Started", tgn, tgn, tgn, tgn)
	} else {
		cmd = fmt.Sprintf("crm conf group gvip%s vip_prtblk_on%s_1 vip_prtblk_on%s_2 vip%s_1 vip%s_2 target%s "+
			"meta target-role=Started", tgn, tgn, tgn, tgn, tgn, tgn)
	}
	out, err := SshCmd(sc, cmd)
	if err != nil {
		errInfo := fmt.Sprintf(strings.Replace(strings.TrimSpace(out), "\n", "", -1))
		Message := client.ApiCallError{client.ApiCallRc{Message: errInfo}}
		return Message
	} else {
		for less := range nodeLess {
			cmd = fmt.Sprintf("crm conf location lo_gvip%s gvip%s -100: %s", tgn, tgn, strconv.Itoa(less))
			if err != nil {
				errInfo := fmt.Sprintf(strings.Replace(strings.TrimSpace(out), "\n", "", -1))
				Message := client.ApiCallError{client.ApiCallRc{Message: errInfo}}
				return Message
			}
		}
	}
	return nil
}

func CreatePortBlockOff(vips []string, tgn string) error {
	sc, _ := GetIPAndConnect(22)
	for vip := range vips {
		cmd := fmt.Sprintf("primitive vip_prtblk_off%s_1 portblock "+
			"ip=%s portno=3260 protocol=tcp action=unblock "+
			"op start timeout=20 interval=0 "+
			"op stop timeout=20 interval=0 "+
			"op monitor timeout=20 interval=20", tgn, strconv.Itoa(vip))
		out, err := SshCmd(sc, cmd)
		if err != nil {
			errInfo := fmt.Sprintf(strings.Replace(strings.TrimSpace(out), "\n", "", -1))
			Message := client.ApiCallError{client.ApiCallRc{Message: errInfo}}
			return Message
		}
	}

	return nil
}

func CreateResourceBond(tgn string) error {
	sc, _ := GetIPAndConnect(22)

	cmd := fmt.Sprintf("crm conf colocation co_prtblkoff%s_1 inf: vip_prtblk_off%s_1 gvip%s", tgn, tgn, tgn)
	out, err := SshCmd(sc, cmd)
	if err != nil {
		errInfo := fmt.Sprintf(strings.Replace(strings.TrimSpace(out), "\n", "", -1))
		Message := client.ApiCallError{client.ApiCallRc{Message: errInfo}}
		return Message
	}

	return nil
}

func CreateNodeAway(ctx context.Context, c *client.Client, tgn string, nodeRun []string) error {
	data := GetNodeData(ctx, c)
	var nodes []string
	for _, node := range data {
		nodes = append(nodes, node["name"])
	}
	nodeAwayList := make([]string, 0)
	for _, node := range nodes {
		found := false
		for _, run := range nodeRun {
			if node == run {
				found = true
				break
			}
		}
		if !found {
			nodeAwayList = append(nodeAwayList, node)
		}
	}
	if len(nodeAwayList) == 0 {

	} else {
		sc, _ := GetIPAndConnect(22)
		for nodeAway := range nodeAwayList {
			cmd := fmt.Sprintf("crm conf location lo_gvip%s_%s gvip%s -inf: %s\n", tgn, strconv.Itoa(nodeAway),
				tgn, strconv.Itoa(nodeAway))
			out, err := SshCmd(sc, cmd)
			if err != nil {
				errInfo := fmt.Sprintf(strings.Replace(strings.TrimSpace(out), "\n", "", -1))
				Message := client.ApiCallError{client.ApiCallRc{Message: errInfo}}
				return Message
			}

		}
	}
	return nil
}

func CreateNodeOn(tgn string, nodeOn string) error {
	sc, _ := GetIPAndConnect(22)
	cmd := fmt.Sprintf("crm res move gvip%s %s", tgn, nodeOn)
	out, err := SshCmd(sc, cmd)
	if err != nil {
		errInfo := fmt.Sprintf(strings.Replace(strings.TrimSpace(out), "\n", "", -1))
		Message := client.ApiCallError{client.ApiCallRc{Message: errInfo}}
		return Message
	}

	return nil
}
