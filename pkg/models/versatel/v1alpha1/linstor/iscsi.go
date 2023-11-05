package linstor

import (
	"context"
	"fmt"
	"github.com/LINBIT/golinstor/client"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type Target struct {
	Name          string   `yaml:"name"`
	Iqn           string   `yaml:"iqn"`
	Tgn           int      `yaml:"tgn"`
	Vip           []string `yaml:"vip"`
	RunNode       []string `yaml:"run_node"`
	AssistantNode []string `yaml:"assistant_node"`
	StartNode     string   `yaml:"start_node"`
	Lun           []string `yaml:"lun"`
}

type Node struct {
	hostname string `yaml:"hostName"`
	iqn      string `yaml:"iqn"`
}

type Lun struct {
	lun    string   `yaml:"lun"`
	number int      `yaml:"number"`
	host   []string `yaml:"host"`
}

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

func Registered(hostname string, iqn string) error {

	var nodes []Node
	data, err := ioutil.ReadFile("/etc/linstorip/host.yaml")
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatalf("error: %v", err)
			return err
		}
	} else {
		err = yaml.Unmarshal(data, &nodes)
		if err != nil {
			log.Fatalf("error: %v", err)
			return err
		}
	}

	nodes = append(nodes, Node{hostname: hostname, iqn: iqn})

	data, err = yaml.Marshal(&nodes)
	if err != nil {
		log.Fatalf("error: %v", err)
		return err
	}

	// Write the data back to the file
	err = ioutil.WriteFile("/etc/linstorip/host.yaml", data, 0644)
	if err != nil {
		log.Fatalf("error: %v", err)
		return err
	}

	return nil
}
func GetRegistered() []map[string]string {
	var nodes []map[string]string

	data, err := ioutil.ReadFile("/etc/linstorip/host.yaml")
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatalf("err：%v", err)
		}
		return nodes
	}

	err = yaml.Unmarshal(data, &nodes)
	if err != nil {
		log.Fatalf("err：%v", err)
		return nodes
	}

	return nodes
}
func SaveTarget(name string, iqn string, tng int, vip []string, NodeRun []string, NodeLess []string,
	NodeOn string, lun []string) error {
	var targets []Target
	data, err := ioutil.ReadFile("/etc/linstorip/target.yaml")
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatalf("error: %v", err)
			return err
		}
	} else {
		err = yaml.Unmarshal(data, &targets)
		if err != nil {
			log.Fatalf("error: %v", err)
			return err
		}
	}

	newTarget := Target{
		Name:          name,
		Iqn:           iqn,
		Tgn:           tng,
		Vip:           vip,
		RunNode:       NodeRun,
		AssistantNode: NodeLess,
		StartNode:     NodeOn,
		Lun:           lun,
	}
	targets = append(targets, newTarget)

	data, err = yaml.Marshal(&targets)
	if err != nil {
		log.Fatalf("error: %v", err)
		return err
	}

	err = ioutil.WriteFile("/etc/linstorip/target.yaml", data, 0644)
	if err != nil {
		log.Fatalf("error: %v", err)
		return err
	}

	return nil
}

func GetTgn() (int, error) {
	filePath := "/etc/linstorip/target.yaml"

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return 1, nil
	}

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return 0, err
	}

	var tgnList []Target
	err = yaml.Unmarshal(data, &tgnList)
	if err != nil {
		return 0, err
	}

	tgn := 1
	for _, target := range tgnList {
		if contain(target.Tgn, tgn) {
			tgn++
		}
	}

	return tgn, nil
}

func contain(list int, item int) bool {
	if list == item {
		return true
	}
	return false
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

func CreateNodeAway(ctx context.Context, c *client.Client, tgn string, nodeRun []string, nodeLess []string) error {
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
		for _, less := range nodeLess {
			if node == less {
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

func ShowTarget() []map[string]string {
	var targets []Target
	var result []map[string]string
	data, err := ioutil.ReadFile("/etc/linstorip/target.yaml")
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return result
	}
	err = yaml.Unmarshal(data, &targets)
	if err != nil {
		return nil
	}

	for _, target := range targets {
		targetMap := map[string]string{
			"name":        target.Name,
			"storageNum":  strconv.Itoa(len(target.Lun)),
			"storageList": strings.Join(target.Lun, ","),
			"vipList":     strings.Join(target.Vip, ","),
		}
		result = append(result, targetMap)
	}
	return result
}

func DeleteTarget(tgn string, vips []string) error {
	sc, _ := GetIPAndConnect(22)
	cmd := fmt.Sprintf("crm res constraints target%s", tgn)
	out, err := SshCmd(sc, cmd)
	if err != nil {
		errInfo := fmt.Sprintf(strings.Replace(strings.TrimSpace(out), "\n", "", -1))
		Message := client.ApiCallError{client.ApiCallRc{Message: errInfo}}
		return Message
	} else if strings.Contains(out, "drbd") {
		Message := client.ApiCallError{client.ApiCallRc{Message: "Target绑定了DRBD资源，不能删除该Target"}}
		return Message
	} else {
		if len(vips) == 1 {
			cmd = fmt.Sprintf("crm conf delete vip_prtblk_on%s_1 vip%s_1 target%s vip_prtblk_off%s_1 "+
				"--force", tgn, tgn, tgn, tgn)
		} else {
			cmd = fmt.Sprintf("crm conf delete vip_prtblk_on%s_1 vip_prtblk_on%s_2"+
				" vip%s_1 vip%s_2 target%s vip_prtblk_off%s_1 vip_prtblk_off%s_2 --force",
				tgn, tgn, tgn, tgn, tgn, tgn, tgn)
		}
		out, err := SshCmd(sc, cmd)
		if err != nil {
			errInfo := fmt.Sprintf(strings.Replace(strings.TrimSpace(out), "\n", "", -1))
			Message := client.ApiCallError{client.ApiCallRc{Message: errInfo}}
			return Message
		}
	}

	return nil
}

func FindTargetOfName(name string) (*Target, error) {
	var targets []Target
	data, err := ioutil.ReadFile("/etc/linstorip/target.yaml")
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, &targets)
	if err != nil {
		return nil, err
	}
	for _, target := range targets {
		if target.Name == name {
			return &target, nil
		}
	}
	return nil, fmt.Errorf("target with name %s not found", name)
}

func ConfigureDRBD(ctx context.Context, c *client.Client, target *Target, resName string) error {
	nodeRun := target.RunNode
	nodeLess := target.AssistantNode
	cloneMax := len(nodeRun) + len(nodeLess)
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
		for _, less := range nodeLess {
			if node == less {
				found = true
				break
			}
		}
		if !found {
			nodeAwayList = append(nodeAwayList, node)
		}
	}

	sc, err := GetIPAndConnect(22)
	if err != nil {
		return err
	}

	cmd := fmt.Sprintf("primitive p_drbd_%s ocf:linbit:drbd "+
		"params drbd_resource=%s "+
		"op monitor interval=29 role=Master "+
		"op monitor interval=30 role=Slave "+
		"ms ms_drbd_%s p_drbd_%s "+
		"meta master-max=1 master-node-max=1 clone-max=%s clone-node-max=1 notify=true target-role=Started ",
		resName, resName, resName, resName, strconv.Itoa(cloneMax))

	out, err := SshCmd(sc, cmd)
	if err != nil {
		errInfo := strings.TrimSpace(out)
		return client.ApiCallError{client.ApiCallRc{Message: errInfo}}
	}

	if len(nodeAwayList) > 0 {
		for _, away := range nodeAwayList {
			cmd := fmt.Sprintf("location DRBD_%s_%s ms_drbd_%s -inf: %s",
				resName, away, resName, away)

			out, err := SshCmd(sc, cmd)
			if err != nil {
				errInfo := strings.TrimSpace(out)
				return client.ApiCallError{client.ApiCallRc{Message: errInfo}}
			}

			target.Lun = append(target.Lun, resName)
			data, err := yaml.Marshal(target)
			if err != nil {
				return err
			}

			err = ioutil.WriteFile("/etc/linstorip/target.yaml", data, 0644)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
func DeleteDRBD(resName string) error {
	sc, _ := GetIPAndConnect(22)
	cmd := fmt.Sprintf("crm res constraints p_drbd_%s", resName)
	out, err := SshCmd(sc, cmd)
	if err != nil {
		errInfo := fmt.Sprintf(strings.Replace(strings.TrimSpace(out), "\n", "", -1))
		Message := client.ApiCallError{client.ApiCallRc{Message: errInfo}}
		return Message
	} else if strings.Contains(out, "LUN") {
		Message := client.ApiCallError{client.ApiCallRc{Message: "该DRBD资源已经被映射，不能删除"}}
		return Message
	} else {
		cmd := fmt.Sprintf("crm conf delete p_drbd_%s --force", resName)
		out, err := SshCmd(sc, cmd)
		if err != nil {
			errInfo := fmt.Sprintf(strings.Replace(strings.TrimSpace(out), "\n", "", -1))
			Message := client.ApiCallError{client.ApiCallRc{Message: errInfo}}
			return Message
		}
	}

	return nil
}

func ShowDRBD() error {
	sc, _ := GetIPAndConnect(22)
	cmd := fmt.Sprintf("crm conf show \"p_drbd_*\" | grep \"drbd_resource=\" | grep -oP '(?<=\\=).*?(?=\\\\)' | grep -v \"linstordb\"")
	out, err := SshCmd(sc, cmd)
	if err != nil {
		errInfo := fmt.Sprintf(strings.Replace(strings.TrimSpace(out), "\n", "", -1))
		Message := client.ApiCallError{client.ApiCallRc{Message: errInfo}}
		return Message
	}

	return nil
}

func findTargetOfRes(resName string) (*Target, error) {
	var targets []Target
	data, err := ioutil.ReadFile("/etc/linstorip/target.yaml")
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, &targets)
	if err != nil {
		return nil, err
	}
	for _, target := range targets {
		for _, lun := range target.Lun {
			if lun == resName {
				return &target, nil
			}
		}
	}
	return nil, fmt.Errorf("lun with resource %s not found", resName)
}
func GetNum() (int, error) {
	filePath := "/etc/linstorip/lun.yaml"

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return 1, nil
	}

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return 0, err
	}

	var numList []Lun
	err = yaml.Unmarshal(data, &numList)
	if err != nil {
		return 0, err
	}

	number := 1
	for _, num := range numList {
		if contain(num.number, number) {
			number++
		}
	}

	return number, nil
}
func CreateISCSI(target Target, node Node, unMap string, resName string, number string) error {
	sc, _ := GetIPAndConnect(22)
	cmd := fmt.Sprintf("linstor r list-volumes -r %s | awk 'NR>2 {print $12}' | head -n 2", resName)
	Device, err := SshCmd(sc, cmd)
	if err != nil {
		errInfo := fmt.Sprintf(strings.Replace(strings.TrimSpace(Device), "\n", "", -1))
		Message := client.ApiCallError{client.ApiCallRc{Message: errInfo}}
		return Message
	} else {
		cmd := fmt.Sprintf("crm conf show LUN_%s", resName)
		out, err := SshCmd(sc, cmd)
		if err != nil {
			errInfo := fmt.Sprintf(strings.Replace(strings.TrimSpace(out), "\n", "", -1))
			Message := client.ApiCallError{client.ApiCallRc{Message: errInfo}}
			return Message
		} else {
			if strings.Contains(out, "not") {
				if unMap == "1" {
					cmd = fmt.Sprintf("crm conf primitive LUN_%s iSCSILogicalUnit "+
						"params target_iqn=\"%s\" implementation=lio-t lun=%s path=\"%s\" emulate_tpu=1 allowed_initiators=\"%s\" "+
						"op start timeout=40 interval=0 "+
						"op stop timeout=40 interval=0 "+
						"op monitor timeout=40 interval=15 "+
						"meta target-role=Stopped", resName, target.Iqn, number, Device, node.iqn)
					SshCmd(sc, cmd)
				} else {
					cmd = fmt.Sprintf("crm conf primitive LUN_%s iSCSILogicalUnit "+
						"params target_iqn=\"%s\" implementation=lio-t lun=%s path=\"%s\" emulate_tpu=0 allowed_initiators=\"%s\" "+
						"op start timeout=40 interval=0 "+
						"op stop timeout=40 interval=0 "+
						"op monitor timeout=40 interval=15 "+
						"meta target-role=Stopped", resName, target.Iqn, number, Device, node.iqn)
					SshCmd(sc, cmd)
				}
				if len(target.Vip) == 1 {

				}
			}
		}
	}

	return nil
}
