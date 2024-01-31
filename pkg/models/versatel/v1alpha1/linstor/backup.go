package linstor

import (
	"context"
	"fmt"
	"github.com/LINBIT/golinstor/client"
	"regexp"
	"strings"
	"time"
)

func GetRemoteData(ctx context.Context, c *client.Client) []map[string]interface{} {
	remotes, err := c.Remote.GetAllLinstor(ctx)
	var remotesInfo []map[string]interface{}
	if err != nil {
		errMap := map[string]interface{}{
			"error": err.Error(),
		}
		return []map[string]interface{}{errMap}
	}
	for _, remote := range remotes {
		remoteInfo := map[string]interface{}{
			"remoteName": remote.RemoteName,
			"url":        remote.Url,
		}
		remotesInfo = append(remotesInfo, remoteInfo)
	}
	return remotesInfo
}

func CreateRemote(ctx context.Context, c *client.Client, remoteName, url, clusterId string) error {
	remote := client.LinstorRemote{
		RemoteName: remoteName,
		Url:        url,
		Passphrase: "",
		ClusterId:  clusterId,
	}
	err := c.Remote.CreateLinstor(ctx, remote)
	return err
}

func DeleteRemote(ctx context.Context, c *client.Client, remoteName string) error {
	err := c.Remote.Delete(ctx, remoteName)
	return err
}

func CreateShip(ctx context.Context, c *client.Client, remoteName string, resNames []string) error {
	var err error
	for _, resName := range resNames {
		dstResName := resName + "_" + time.Now().Format("20060102150405")
		ship := client.BackupShipRequest{
			SrcRscName: resName,
			DstRscName: dstResName,
		}
		err = c.Backup.Ship(ctx, remoteName, ship)
		if err != nil {
			return err
		}
	}
	return err
}

func GetScheduleData() []map[string]interface{} {
	sc, _ := GetIPAndConnect(22)
	cmd := "linstor schedule list | awk 'BEGIN{FS=\"|\"} NR>2 {print $2}';" +
		"linstor schedule list | awk 'BEGIN{FS=\"|\"} NR>2 {print $4}';" +
		"linstor schedule list | awk 'BEGIN{FS=\"|\"} NR>2 {print $5}';" +
		"linstor schedule list | awk 'BEGIN{FS=\"|\"} NR>2 {print $6}';" +
		"linstor schedule list | awk 'BEGIN{FS=\"|\"} NR>2 {print $7}'"
	out, err := SshCmd(sc, cmd)
	if err != nil {
		errInfo := fmt.Sprintf(strings.Replace(strings.TrimSpace(out), "\n", "", -1))
		Message := client.ApiCallError{client.ApiCallRc{Message: errInfo}}
		return []map[string]interface{}{{"error": Message}}
	}

	lines := strings.Split(out, "\n")
	numSchedules := len(lines) / 5

	var scheduleData []map[string]interface{}
	for i := 0; i < numSchedules; i++ {
		scheduleInfo := map[string]interface{}{
			"scheduleName": strings.TrimSpace(lines[i]),
			"incremental":  strings.TrimSpace(lines[i+numSchedules]),
			"keepLocal":    strings.TrimSpace(lines[i+2*numSchedules]),
			"keepRemote":   strings.TrimSpace(lines[i+3*numSchedules]),
			"onFailure":    strings.TrimSpace(lines[i+4*numSchedules]),
		}

		if isValidScheduleInfo(scheduleInfo) {
			scheduleData = append(scheduleData, scheduleInfo)
		}
	}
	return scheduleData
}

func isValidScheduleInfo(scheduleInfo map[string]interface{}) bool {
	for _, value := range scheduleInfo {
		strValue := value.(string)
		if strValue != "" && !strings.Contains(strValue, "===") {
			return true
		}
	}
	return false
}

func CreateSchedule(scheduleName, incremental, keepLocal, keepRemote, onFailure string) error {
	sc, _ := GetIPAndConnect(22)
	cmd := fmt.Sprintf("linstor schedule create --incremental-cron '%s' --keep-local %s --keep-remote %s --on-failure RETRY "+
		"--max-retries %s %s", incremental, keepLocal, keepRemote, onFailure, scheduleName)
	out, err := SshCmd(sc, cmd)
	if err != nil {
		errInfo := fmt.Sprintf(strings.Replace(strings.TrimSpace(out), "\n", "", -1))
		Message := client.ApiCallError{client.ApiCallRc{Message: errInfo}}
		return Message
	}
	return nil
}

func DeleteSchedule(scheduleName string) error {
	sc, _ := GetIPAndConnect(22)
	cmd := fmt.Sprintf("linstor schedule delete %s", scheduleName)
	out, err := SshCmd(sc, cmd)
	if err != nil {
		errInfo := fmt.Sprintf(strings.Replace(strings.TrimSpace(out), "\n", "", -1))
		Message := client.ApiCallError{client.ApiCallRc{Message: errInfo}}
		return Message
	}
	return nil
}

func GetBackupData() []map[string]interface{} {
	sc, _ := GetIPAndConnect(22)
	cmd := "linstor schedule list-by-resource --active-only"
	out, err := SshCmd(sc, cmd)
	if err != nil {
		errInfo := fmt.Sprintf(strings.Replace(strings.TrimSpace(out), "\n", "", -1))
		Message := client.ApiCallError{client.ApiCallRc{Message: errInfo}}
		return []map[string]interface{}{{"error": Message}}
	}
	re := regexp.MustCompile(`\|\s*([^|]*?)\s*\|\s*([^|]*?)\s*\|\s*([^|]*?)\s*\|\s*([^|]*?)\s*\|\s*([^|]*?)\s*\|\s*([^|]*?)\s*\|\s*([^|]*?)\s*\|\s*([^|]*?)\s*\|`)
	matches := re.FindAllStringSubmatch(out, -1)

	var backupData []map[string]interface{}
	for _, match := range matches {
		if len(match) == 9 {
			backupInfo := map[string]interface{}{
				"resName":      match[1],
				"remoteName":   match[2],
				"scheduleName": match[3],
				"lastPlan":     match[4],
				"nextPlan":     match[5],
			}
			backupData = append(backupData, backupInfo)
		}
	}
	return backupData
}

func CreateBackup(resName, remoteName, scheduleName string) error {
	sc, _ := GetIPAndConnect(22)
	cmd := fmt.Sprintf("linstor schedule backup schedule enable --rd %s %s %s", resName, remoteName, scheduleName)
	out, err := SshCmd(sc, cmd)
	if err != nil {
		errInfo := fmt.Sprintf(strings.Replace(strings.TrimSpace(out), "\n", "", -1))
		Message := client.ApiCallError{client.ApiCallRc{Message: errInfo}}
		return Message
	}
	return nil
}

func DeleteBackup(resName, remoteName, scheduleName string) error {
	sc, _ := GetIPAndConnect(22)
	cmd := fmt.Sprintf("linstor backup schedule delete  --rd %s %s %s", resName, remoteName, scheduleName)
	out, err := SshCmd(sc, cmd)
	if err != nil {
		errInfo := fmt.Sprintf(strings.Replace(strings.TrimSpace(out), "\n", "", -1))
		Message := client.ApiCallError{client.ApiCallRc{Message: errInfo}}
		return Message
	}
	return nil
}

func GetClusterId(ctx context.Context, c *client.Client) []map[string]interface{} {
	propsInfos, err := c.Controller.GetProps(ctx)
	if err != nil {
		errMap := map[string]interface{}{
			"error": err.Error(),
		}
		return []map[string]interface{}{errMap}
	}
	var datas []map[string]interface{}
	localID, _ := propsInfos["Cluster/LocalID"]
	data := map[string]interface{}{
		"clusterid": localID,
	}
	datas = append(datas, data)
	return datas

}
