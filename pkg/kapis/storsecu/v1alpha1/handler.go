package v1alpha1

import (
	"fmt"

	"github.com/emicklei/go-restful"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"kubesphere.io/kubesphere/pkg/api"
	//servererr "kubesphere.io/kubesphere/pkg/server/errors"
	//"kubesphere.io/kubesphere/pkg/apiserver/auditing"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	backupv1alpha1 "kubesphere.io/kubesphere/pkg/models/storsecu/v1alpha1/backup"
	servererr "kubesphere.io/kubesphere/pkg/server/errors"
)

//var PyStr = python3.PyUnicode_FromString
//var PyInt = python3.PyLong_FromLong
//var GoStr = python3.PyUnicode_AsUTF8
//var GoInt = python3.PyLong_AsLong

type handler struct {
	Client kubernetes.Interface
	Config *rest.Config
}

func newHandler(client kubernetes.Interface, config *rest.Config) *handler {
	return &handler{Client: client, Config: config}
}

type MessageOP struct {
	Result string `json:"result"`
	Info   string `json:"info"`
}

type MessageList struct {
	Code  int                 `json:"code"`
	Count int                 `json:"count"`
	Data  []map[string]string `json:"data"`
}

type BackupMetadata struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type BackupInfo struct {
	Metadata   BackupMetadata `json:"metadata"`
	Snapshot   string         `json:"snapshot"`
	Vg         string         `json:"vg"`
	Image      string         `json:"image"`
	BackupType string         `json:"backupType"`
	Schedule   string         `json:"schedule"`
}

//func init(){
//	gp.Initialize()
//	gp.ImportSystemModule()
//	gp.ImportCustomModule("/home/samba/kubesphere.io/kubesphere/vplx")
//	gp.ImportCustomModule("/home/samba/kubesphere.io/kubesphere/pythoncode")

func (h *handler) handleListbackupres(req *restful.Request, resp *restful.Response) {
	query := query.ParseQueryParameter(req)
	resname := "test"
	data, _ := backupv1alpha1.GetBackupRes(h.Client, h.Config, resname)

	if data != nil {
		message := backupv1alpha1.LinstorGetter{0, len(data), data}
		message.List(query)
		resp.WriteAsJson(message)
	}
}

func (h *handler) Backup(req *restful.Request, resp *restful.Response) {

	backupInfo := new(BackupInfo)
	err := req.ReadEntity(&backupInfo)
	fmt.Println("backupInfo: ", backupInfo)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	if backupInfo.Metadata.Type == "snapshot" {

		err = backupv1alpha1.ToSnapShot(h.Client, h.Config, backupInfo.Metadata.Name)

	} else if backupInfo.Metadata.Type == "image" {

		err = backupv1alpha1.ToImage(h.Client, h.Config, backupInfo.Metadata.Name, backupInfo.Snapshot)

	} else if backupInfo.Metadata.Type == "snapshotRestore" {

		err = backupv1alpha1.SnapshotRestore(h.Client, h.Config, backupInfo.Metadata.Name, backupInfo.Snapshot)

	} else if backupInfo.Metadata.Type == "imageRestore" {

		err = backupv1alpha1.ImageRestore(h.Client, h.Config, backupInfo.Metadata.Name, backupInfo.Image, backupInfo.Vg)

	} else if backupInfo.Metadata.Type == "time" {

		err = backupv1alpha1.RegularBackup(h.Client, h.Config, backupInfo.Metadata.Name, backupInfo.Schedule, backupInfo.BackupType)

	}

	if err != nil {
		resp.WriteAsJson(err)
		return
	}

	resp.WriteEntity(servererr.None)

}

func (h *handler) AntiVirus(req *restful.Request, resp *restful.Response) {

	backupInfo := new(BackupInfo)
	err := req.ReadEntity(&backupInfo)
	fmt.Println("backupInfo: ", backupInfo)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	if backupInfo.Metadata.Type == "snapshot" {

		err = backupv1alpha1.ToSnapShot(h.Client, h.Config, backupInfo.Metadata.Name)

	} else if backupInfo.Metadata.Type == "image" {

		err = backupv1alpha1.ToImage(h.Client, h.Config, backupInfo.Metadata.Name, backupInfo.Snapshot)

	} else if backupInfo.Metadata.Type == "snapshotRestore" {

		err = backupv1alpha1.SnapshotRestore(h.Client, h.Config, backupInfo.Metadata.Name, backupInfo.Snapshot)

	} else if backupInfo.Metadata.Type == "imageRestore" {

		err = backupv1alpha1.ImageRestore(h.Client, h.Config, backupInfo.Metadata.Name, backupInfo.Image, backupInfo.Vg)

	} else if backupInfo.Metadata.Type == "time" {

		err = backupv1alpha1.RegularBackup(h.Client, h.Config, backupInfo.Metadata.Name, backupInfo.Schedule, backupInfo.BackupType)

	}

	if err != nil {
		resp.WriteAsJson(err)
		return
	}

	resp.WriteEntity(servererr.None)

}
