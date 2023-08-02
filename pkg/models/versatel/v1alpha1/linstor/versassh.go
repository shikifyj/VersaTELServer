package linstor

import (
	"context"
	"errors"
	"fmt"
	"github.com/LINBIT/golinstor/client"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"math/rand"
	"regexp"
	"strings"
	"time"

	//"path"
	"github.com/dlclark/regexp2"
)

type SshConnect struct {
	Sshclient *ssh.Client
	Host      string
}

type SshConnectList struct {
	Connects []*SshConnect
}

var sshlist = &SshConnectList{}

func SSHConnect(user, host string, port int) (*ssh.Client, error) {
	var (
		addr         string
		clientConfig *ssh.ClientConfig
		client       *ssh.Client
		err          error
	)

	//homePath, err := os.UserHomeDir()
	//if err != nil {
	//  return nil, err
	//}
	//key, err := ioutil.ReadFile(path.Join(homePath, ".ssh", "id_rsa"))
	key, err := ioutil.ReadFile("/etc/localssh/id_rsa")
	//key, err := ioutil.ReadFile("/root/.ssh/id_rsa")
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}

	clientConfig = &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		Timeout:         30 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// connet to ssh
	addr = fmt.Sprintf("%s:%d", host, port)

	if client, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}
	return client, nil
}

func GetSshList(ctx context.Context, c *client.Client) {

	if sshlist.Connects == nil {
		DoSshs(ctx, c)
	}
}

func DoSshs(ctx context.Context, c *client.Client) {

	data := GetNodesIP(ctx, c)

	for _, node := range data {
		reg := regexp.MustCompile(`\d+\.\d+\.\d+\.\d+`)
		result := reg.FindAllStringSubmatch(node["addr"], -1)
		if result[0][0] != "10.203.1.240" {
			sshclient, err := SSHConnect("root", result[0][0], 22)
			if err != nil {
				sshlist.Connects = nil
				log.Fatal(err)
			}
			sc := &SshConnect{sshclient, node["name"]}
			sshlist.Connects = append(sshlist.Connects, sc)
		}
	}
}

func SshCmd(sshclient *ssh.Client, cmd string) (string, error) {

	session, err := sshclient.NewSession()
	if err != nil {
		sshlist.Connects = nil
		fmt.Println("Error creating ssh session", err)
	}
	defer session.Close()
	//执行远程命令
	combo, err := session.CombinedOutput(cmd)
	if err != nil {
		sshlist.Connects = nil
		fmt.Println("Error cmd to linstor node: ", string(combo))
	}
	return string(combo), err
}

func GetNodesIP(ctx context.Context, c *client.Client) []map[string]string {

	nodesInfo := []map[string]string{}

	data := GetNodeData(ctx, c)

	for _, node := range data {

		nodeInfo := map[string]string{
			"name": node["name"],
			"addr": node["addr"],
		}
		nodesInfo = append(nodesInfo, nodeInfo)

	}
	return nodesInfo
}

func CreatePV(ctx context.Context, c *client.Client, devName string, nodeName string) error {

	var err error = errors.New("Can not find the node: " + nodeName)
	GetSshList(ctx, c)
	//clusterPV := []map[string]string{}
	for _, cli := range sshlist.Connects {

		if nodeName == cli.Host {
			err = nil
			out, err := SshCmd(cli.Sshclient, "pvcreate "+devName)
			if err != nil {
				errInfo := fmt.Sprintf(strings.Replace(strings.TrimSpace(out), "\n", "", -1))
				Message := client.ApiCallError{client.ApiCallRc{Message: errInfo}}
				return Message
			}

			break
		}
	}
	return err

}

func CreateVG(ctx context.Context, c *client.Client, pvName []string, vgName string, nodeName string) error {

	var err error = errors.New("Can not find the node: " + nodeName)
	GetSshList(ctx, c)
	//clusterPV := []map[string]string{}
	for _, cli := range sshlist.Connects {

		if nodeName == cli.Host {
			err = nil
			out, err := SshCmd(cli.Sshclient, "vgcreate "+vgName+" "+strings.Join(pvName, " "))
			if err != nil {
				errInfo := fmt.Sprintf(strings.Replace(strings.TrimSpace(out), "\n", "", -1))
				Message := client.ApiCallError{client.ApiCallRc{Message: errInfo}}
				return Message
			}

			break
		}
	}
	return err

}

func CreateThinPool(ctx context.Context, c *client.Client, size string, thinPoolName string, vgName string, nodeName string) error {

	var err error = errors.New("Can not find the node: " + nodeName)
	GetSshList(ctx, c)
	//clusterPV := []map[string]string{}
	for _, cli := range sshlist.Connects {

		if nodeName == cli.Host {
			err = nil
			out, err := SshCmd(cli.Sshclient, "lvcreate -L "+size+" --thinpool "+thinPoolName+" "+vgName)
			if err != nil {
				errInfo := fmt.Sprintf(strings.Replace(strings.TrimSpace(out), "\n", "", -1))
				Message := client.ApiCallError{client.ApiCallRc{Message: errInfo}}
				return Message
			}

			break
		}
	}
	return err

}

func CreateDeviceThinPool(ctx context.Context, c *client.Client, size string, thinPoolName string, Device []string, nodeName string) error {

	var err error = errors.New("Can not find the node: " + nodeName)
	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(100) + 1
	vgName := fmt.Sprintf("vvg_%s_%d", thinPoolName, randomNumber)
	GetSshList(ctx, c)
	for _, cli := range sshlist.Connects {
		if nodeName == cli.Host {
			err = nil
			for _, Device := range Device {
				out, err := SshCmd(cli.Sshclient, "pvcreate "+Device)
				if err != nil {
					errInfo := fmt.Sprintf(strings.Replace(strings.TrimSpace(out), "\n", "", -1))
					Message := client.ApiCallError{client.ApiCallRc{Message: errInfo}}
					return Message
				}
			}
			out, err := SshCmd(cli.Sshclient, "vgcreate "+vgName+" "+strings.Join(Device, " "))
			if err != nil {
				errInfo := fmt.Sprintf(strings.Replace(strings.TrimSpace(out), "\n", "", -1))
				Message := client.ApiCallError{client.ApiCallRc{Message: errInfo}}
				return Message
			}
			out, err = SshCmd(cli.Sshclient, "lvcreate -L "+size+" --thinpool "+thinPoolName+" "+vgName)
			if err != nil {
				errInfo := fmt.Sprintf(strings.Replace(strings.TrimSpace(out), "\n", "", -1))
				Message := client.ApiCallError{client.ApiCallRc{Message: errInfo}}
				return Message
			}
			break
		}
	}
	return err

}

func CreateLV(ctx context.Context, c *client.Client, size string, lvName string, thinPoolName string, vgName string, nodeName string) error {

	var err error = errors.New("Can not find the node: " + nodeName)
	GetSshList(ctx, c)
	//clusterPV := []map[string]string{}
	for _, cli := range sshlist.Connects {

		if nodeName == cli.Host {
			err = nil
			out, err := SshCmd(cli.Sshclient, "lvcreate -V "+size+" --thin -n "+lvName+" "+vgName+"/"+thinPoolName)
			if err != nil {
				errInfo := fmt.Sprintf(strings.Replace(strings.TrimSpace(out), "\n", "", -1))
				Message := client.ApiCallError{client.ApiCallRc{Message: errInfo}}
				return Message
			}

			break
		}
	}
	return err

}

func GetLvmDevices(ctx context.Context, c *client.Client) []map[string]string {
	GetSshList(ctx, c)
	clusterLvm := []map[string]string{}

	for _, cli := range sshlist.Connects {
		// Execute df command via SSH
		dfOutput, _ := SshCmd(cli.Sshclient, "df")

		// Parse df output to get device names
		dfLines := strings.Split(dfOutput, "\n")
		dfDevices := make(map[string]bool)
		for _, line := range dfLines[1:] {
			parts := strings.Fields(line)
			if len(parts) > 0 && strings.HasPrefix(parts[0], "/dev") {
				dfDevices[parts[0]] = true
			}
		}

		replay, _ := SshCmd(cli.Sshclient, "lvmdiskscan")
		expr := `(?<=\[\D*)\d.+(?=\])`
		regAll := regexp.MustCompile(`/dev/sd.+`)
		regDevice := regexp.MustCompile(`/dev/sd\s*`)
		regSize, _ := regexp2.Compile(expr, 0)

		resultAll := regAll.FindAllStringSubmatch(replay, -1)

		for _, data := range resultAll {
			if !strings.Contains(data[0], "LVM") {
				resultDevice := regDevice.FindAllStringSubmatch(data[0], -1)
				// Check if device is in dfDevices
				if _, ok := dfDevices[resultDevice[0][0]]; !ok {
					m, _ := regSize.FindStringMatch(data[0])
					resultSize := m.String()
					resSize, _ := ParseSizeForLvm(resultSize)

					lvmInfo := map[string]string{
						"node": cli.Host,
						"name": resultDevice[0][0],
						"size": resSize,
					}
					clusterLvm = append(clusterLvm, lvmInfo)
				}
			}
		}
	}
	return clusterLvm
}

func GetLvmPVs(ctx context.Context, c *client.Client) []map[string]string {
	GetSshList(ctx, c)
	clusterPV := []map[string]string{}
	for _, cli := range sshlist.Connects {

		replay, _ := SshCmd(cli.Sshclient, "pvs")
		regAll := regexp.MustCompile(`/dev/sd[^a].+`)
		regPV := regexp.MustCompile(`/dev/sd[^a][^\s]*`)
		regVG, _ := regexp2.Compile(`(?<=\s\D*)[^\s]+`, 0)
		regSize, _ := regexp2.Compile(`(?<=\s\D*)\d+\.\w+`, 0)

		resultAll := regAll.FindAllStringSubmatch(replay, -1)

		for _, data := range resultAll {
			resultPV := regPV.FindAllStringSubmatch(data[0], -1)
			m, _ := regVG.FindStringMatch(data[0])
			resultVG := m.String()
			if resultVG == "lvm2" {
				resultVG = ""
			}
			n, _ := regSize.FindStringMatch(data[0])
			resultSize := n.String()

			lvmInfo := map[string]string{
				"node": cli.Host,
				"name": resultPV[0][0],
				"vg":   resultVG,
				"size": resultSize,
			}
			clusterPV = append(clusterPV, lvmInfo)

		}

	}
	return clusterPV

}

func GetLvmVGs(ctx context.Context, c *client.Client) []map[string]string {
	GetSshList(ctx, c)
	clusterVG := []map[string]string{}
	for _, cli := range sshlist.Connects {

		replay, _ := SshCmd(cli.Sshclient, "vgs")
		regAll := regexp.MustCompile(`[^\n]+`)
		regVG := regexp.MustCompile(`[^\s]+`)
		regSize, _ := regexp2.Compile(`(?<=\s\D*)\d+\.\w+`, 0)

		resultAll := regAll.FindAllStringSubmatch(replay, -1)
		resultAll = resultAll[1:]

		for _, data := range resultAll {

			resultVG := regVG.FindString(data[0])

			n, _ := regSize.FindStringMatch(data[0])
			resultSize := n.String()

			lvmInfo := map[string]string{
				"node": cli.Host,
				"vg":   resultVG,
				"size": resultSize,
			}
			clusterVG = append(clusterVG, lvmInfo)

		}

	}
	return clusterVG

}

func GetLvmLVs(ctx context.Context, c *client.Client) []map[string]string {
	GetSshList(ctx, c)
	clusterLV := []map[string]string{}
	for _, cli := range sshlist.Connects {

		replay, _ := SshCmd(cli.Sshclient, "lvs")
		regAll := regexp.MustCompile(`[^\n]+`)
		regInfo := regexp.MustCompile(`[^\s]+`)

		resultAll := regAll.FindAllStringSubmatch(replay, -1)
		resultAll = resultAll[1:]

		for _, data := range resultAll {

			//pool := ""
			resultInfo := regInfo.FindAllStringSubmatch(data[0], -1)
			if len(resultInfo) < 5 {
				continue
			}
			//if !strings.Contains(resultInfo[2][0], ".") {
			//	pool = resultInfo[4][0]
			//}

			lvmInfo := map[string]string{
				"node": cli.Host,
				"name": resultInfo[0][0],
				"vg":   resultInfo[1][0],
				"size": resultInfo[3][0],
				//			"pool": pool,
			}
			clusterLV = append(clusterLV, lvmInfo)

		}

	}
	return clusterLV

}

func DeletePV(ctx context.Context, c *client.Client, devName string, nodeName string) error {

	var err error = errors.New("Can not find the node: " + nodeName)
	GetSshList(ctx, c)
	//clusterPV := []map[string]string{}
	for _, cli := range sshlist.Connects {

		if nodeName == cli.Host {
			err = nil
			out, err := SshCmd(cli.Sshclient, "pvremove -y "+devName)
			if err != nil {
				errInfo := fmt.Sprintf(strings.Replace(strings.TrimSpace(out), "\n", "", -1))
				Message := client.ApiCallError{client.ApiCallRc{Message: errInfo}}
				return Message
			}

			break
		}
	}
	return err

}

func DeleteVG(ctx context.Context, c *client.Client, vgName string, nodeName string) error {

	var err error = errors.New("Can not find the node: " + nodeName)
	GetSshList(ctx, c)
	//clusterPV := []map[string]string{}
	for _, cli := range sshlist.Connects {

		if nodeName == cli.Host {
			err = nil
			out, err := SshCmd(cli.Sshclient, "vgremove -y "+vgName)
			if err != nil {
				errInfo := fmt.Sprintf(strings.Replace(strings.TrimSpace(out), "\n", "", -1))
				Message := client.ApiCallError{client.ApiCallRc{Message: errInfo}}
				return Message
			}

			break
		}
	}
	return err

}

func DeleteThinPool(ctx context.Context, c *client.Client, thinPoolName string, nodeName string) error {

	var err error = errors.New("Can not find the node: " + nodeName)
	GetSshList(ctx, c)
	//clusterPV := []map[string]string{}
	for _, cli := range sshlist.Connects {

		if nodeName == cli.Host {
			err = nil
			out, err := SshCmd(cli.Sshclient, "lvremove -y "+thinPoolName)
			if err != nil {
				errInfo := fmt.Sprintf(strings.Replace(strings.TrimSpace(out), "\n", "", -1))
				Message := client.ApiCallError{client.ApiCallRc{Message: errInfo}}
				return Message
			}

			break
		}
	}
	return err

}

func DeleteLV(ctx context.Context, c *client.Client, lvName string, nodeName string) error {

	var err error = errors.New("Can not find the node: " + nodeName)
	GetSshList(ctx, c)
	//clusterPV := []map[string]string{}
	for _, cli := range sshlist.Connects {

		if nodeName == cli.Host {
			err = nil
			_, err = SshCmd(cli.Sshclient, "lvremove -y "+lvName)

			break
		}
	}
	return err

}
