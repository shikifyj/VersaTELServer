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
	cmd := "linstor schedule list"
	out, err := SshCmd(sc, cmd)
	if err != nil {
		errInfo := fmt.Sprintf(strings.Replace(strings.TrimSpace(out), "\n", "", -1))
		Message := client.ApiCallError{client.ApiCallRc{Message: errInfo}}
		return []map[string]interface{}{{"error": Message}}
	}
	re := regexp.MustCompile(`\|\s*([^|]+?)\s*\|\s*([^|]+?)\s*\|\s*([^|]+?)\s*\|\s*([^|]+?)\s*\|\s*([^|]+?)\s*\|\s*([^|]+?)\s*\|`)
	matches := re.FindAllStringSubmatch(out, -1)

	var scheduleData []map[string]interface{}
	for _, match := range matches {
		if len(match) == 7 {
			scheduleInfo := map[string]interface{}{
				"scheduleName": match[1],
				"incremental":  match[3],
				"keepLocal":    match[4],
				"keepRemote":   match[5],
				"onFailure":    match[6],
			}
			scheduleData = append(scheduleData, scheduleInfo)
		}
	}
	return scheduleData
}

func CreateSchedule(scheduleName, incremental, keepLocal, keepRemote, onFailure string) error {
	sc, _ := GetIPAndConnect(22)
	cmd := fmt.Sprintf("linstor schedule create --incremental-cron '%s' "+
		"--keep-local %s --keep-remote %s --on-failure RETRY --max-retries %s %s", incremental, keepLocal, keepRemote, onFailure, scheduleName)
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
