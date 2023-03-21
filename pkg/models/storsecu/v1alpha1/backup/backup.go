package backup

import (
	"fmt"

	"bytes"
	"encoding/json"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	//"k8s.io/client-go/tools/clientcmd"
	//"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/tools/remotecommand"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servererr "kubesphere.io/kubesphere/pkg/server/errors"
	"sort"
	"context"
	"strings"
	"regexp"
)


func GetBackupRes(client kubernetes.Interface, config *rest.Config, command string) ([]map[string]string, error) {
	command = "/backup/backup list"
	//command = "/backup/backup"
	data := []map[string]string{}

	strout, err := DoToPod(client, config, command)

	if err != nil {
		fmt.Println(err)
		return data, err
	}

	//strout = "[{'name': 'res_b', 'snapshot': 'resa320230313133527', 'snapshotRestore': '/dev/drbd1001', 'image': 'resa320230313133527.img', 'imageRestore': '/dev/vg/lv02', 'time': '2023-03-13 13:35:27'},{'name': 'thres', 'snapshot': '', 'snapshotRestore': '', 'image': '', 'imageRestore': '', 'time': ''}, {'name': 'thres2', 'snapshot': '', 'snapshotRestore': '', 'image': '', 'imageRestore': '', 'time': ''}]"
	data, err = GetTinResources(strout)
	data = sortSlices(data)

	return data, err
}



func ToSnapShot(client kubernetes.Interface, config *rest.Config, name string) error {
	command := "/backup/backup create -r " + name 

	strout, err := DoToPod(client, config, command)

	if err != nil {
		fmt.Println(err)
		return err
	}

	if strings.Contains(strout, "Success") {
		err = nil
	} else {

		err = servererr.New("Snapshot failded")
	}

	return err
}



func ToImage(client kubernetes.Interface, config *rest.Config, name string, snapshot string) error {
	command := "/backup/backup dump -r " + name + " -s " + snapshot

	strout, err := DoToPod(client, config, command)

	if err != nil {
		fmt.Println(err)
		return err
	}

	if strings.Contains(strout, "Success") {
		err = nil
	} else {

		err = servererr.New("Dump image failded")
	}

	return err
}


func SnapshotRestore(client kubernetes.Interface, config *rest.Config, resname string, snapshot string) error {

	command := "/backup/backup snapshotrestore -r "+resname+" -s "+snapshot
	strout, err := DoToPod(client, config, command)

	if err != nil {
		fmt.Println(err)
		return err
	}

	if strings.Contains(strout, "Success") {
		err = nil
	} else {

		err = servererr.New("SnapshotRestore failded")
	}

	return err
}

func ImageRestore(client kubernetes.Interface, config *rest.Config, resname string, image string, vg string) error {

	command := "/backup/backup imagerestore -r "+resname+" -i "+image+" -vg "+vg
	strout, err := DoToPod(client, config, command)

	if err != nil {
		fmt.Println(err)
		return err
	}

	if strings.Contains(strout, "Success") {
		err = nil
	} else {

		err = servererr.New("ImageRestore failded")
	}

	return err
}


func RegularBackup(client kubernetes.Interface, config *rest.Config, resname string, contab string, backupType string) error {
	kind := "dump"
	if backupType == "snapshot" {
		kind = "create"
	}
	command := "/backup/schedule -s \""+contab+"\" -c \"/backup/backup "+kind+" -r "+resname+"\""

	strout, err := DoToPod(client, config, command)

	if err != nil {
		fmt.Println(err)
		return err
	}

	if strout != "" {
		err = servererr.New("RegularBackup failded")
	}
	return err
}


func DoToPod(client kubernetes.Interface, config *rest.Config, command string) (string, error) {

	fmt.Println("DoToPod: ", command)
	cmd := []string{
		"sh",
		"-c",
		command,
	}


    deploymentName := "feixibackup"
    strout := ""
    pods, err := client.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{
        LabelSelector: "app=" + deploymentName,
    })
    if err != nil {
    	fmt.Println("find feixibackup pod failded")
        return strout, servererr.New("find feixibackup pod failded")
    }

    for _, pod := range pods.Items {

		const tty = false
		req := client.CoreV1().RESTClient().Post().
			Resource("pods").
			Name(pod.Name).
			Namespace("default").SubResource("exec").Param("container", "backup")
		req.VersionedParams(
			&v1.PodExecOptions{
				Command: cmd,
				Stdin:   false,
				Stdout:  true,
				Stderr:  true,
				TTY:     tty,
			},
			scheme.ParameterCodec,
		)

		var stdout, stderr bytes.Buffer
		exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
		if err != nil {
			fmt.Println("remotecommand.NewSPDYExecutor err!!!!")
			continue
		}
		err = exec.Stream(remotecommand.StreamOptions{
			Stdin:  nil,
			Stdout: &stdout,
			Stderr: &stderr,
		})
		if err != nil {
			fmt.Println("exec.Stream err!!!!")
			continue
		}

		strout = strings.TrimSpace(stdout.String())

		fmt.Println("strout: ", strout)
		break

    }

	return strout, err
}


func GetTinResources(s string) ([]map[string]string, error) {
    // 匹配字符串中的 JSON 数据部分
    re := regexp.MustCompile(`\[{.*}\]`)
    match := re.FindStringSubmatch(s)
    if len(match) == 0 {
        return nil, fmt.Errorf("failed to match JSON data")
    }
    jsonData := match[0]

    // 替换字符串中的单引号为双引号，然后解析为 JSON
    jsonData = strings.ReplaceAll(jsonData, "'", "\"")
    var data []map[string]string
    err := json.Unmarshal([]byte(jsonData), &data)
    if err != nil {
        return nil, err
    }
    return data, nil
}


func sortSlices(slices []map[string]string) []map[string]string {
    sort.Slice(slices, func(i, j int) bool {
        if slices[i]["time"] == "" && slices[j]["time"] != "" {
            return false
        } else if slices[i]["time"] != "" && slices[j]["time"] == "" {
            return true
        } else if slices[i]["time"] == "" && slices[j]["time"] == "" {
            return slices[i]["name"] < slices[j]["name"]
        }
        return slices[i]["time"] < slices[j]["time"]
    })
    return slices
}










