package v1alpha1

import (
	"fmt"

	"github.com/emicklei/go-restful"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/auditing"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	linstorv1alpha1 "kubesphere.io/kubesphere/pkg/models/versatel/v1alpha1/linstor"
	servererr "kubesphere.io/kubesphere/pkg/server/errors"
)

//var PyStr = python3.PyUnicode_FromString
//var PyInt = python3.PyLong_FromLong
//var GoStr = python3.PyUnicode_AsUTF8
//var GoInt = python3.PyLong_AsLong

type handler struct {
	//linstorGetterV1alpha1  *linstorv1alpha1.linstorGetter
	ControllerIP string
}

func newHandler(ip string) handler {
	return handler{ip}
}

type MessageList struct {
	Code  int                 `json:"code"`
	Count int                 `json:"count"`
	Data  []map[string]string `json:"data"`
}

type MessageOP struct {
	Result string `json:"result"`
	Info   string `json:"info"`
}

type MessageExist struct {
	Exist bool `json:"exist"`
}

type LinstorNode struct {
	Name     string `json:"name"`
	IP       string `json:"addr"`
	NodeType string `json:"node_type"`
}

type LinstorSP struct {
	Metadata map[string]interface{} `json:"metadata"`
	NodeName string                 `json:"node"`
	Type     string                 `json:"type"`
	Volume   string                 `json:"volume"`
}

type LinstorRes struct {
	Metadata    map[string]interface{} `json:"metadata"`
	Node        []string               `json:"node"`
	StoragePool [][]string             `json:"storagepool"`
	Size        string                 `json:"size"`
}

type LvmPV struct {
	Name string `json:"name"`
	Node string `json:"node"`
}

type LvmVG struct {
	Name string `json:"name"`
	Node string `json:"node"`
	PV   string `json:"pv"`
}

type LvmThinPool struct {
	Name string `json:"name"`
	Node string `json:"node"`
	VG   string `json:"vg"`
	Size string `json:"size"`
}

type LvmLV struct {
	Name     string `json:"name"`
	Node     string `json:"node"`
	VG       string `json:"vg"`
	Size     string `json:"size"`
	ThinPool string `json:"thinpool"`
}

type URLResponse struct {
	URL string `json:"URL"`
}

type DiskfulSP struct {
	NodeName string `json:"nodename"`
}

type ReplicaRes struct {
	ResName         string   `json:"resname"`
	NodeName        []string `json:"nodename"`
	PoolName        []string `json:"poolname"`
	TargetReplicas  int      `json:"originalnum"`
	CurrentReplicas int      `json:"newnum"`
}

//func init(){
//	gp.Initialize()
//	gp.ImportSystemModule()
//	gp.ImportCustomModule("/home/samba/kubesphere.io/kubesphere/vplx")
//	gp.ImportCustomModule("/home/samba/kubesphere.io/kubesphere/pythoncode")
//}

func (h *handler) handleListNodes(req *restful.Request, resp *restful.Response) {
	query := query.ParseQueryParameter(req)
	client, ctx := linstorv1alpha1.GetClient(h.ControllerIP)
	data := linstorv1alpha1.GetNodeData(ctx, client)
	message := linstorv1alpha1.LinstorGetter{0, len(data), data}
	message.List(query)
	resp.WriteAsJson(message)
}

func (h *handler) DescribeNode(req *restful.Request, resp *restful.Response) {
	nodename := req.PathParameter("node")
	client, ctx := linstorv1alpha1.GetClient(h.ControllerIP)
	err := linstorv1alpha1.DescribeNode(ctx, client, nodename)
	if err != nil {
		resp.WriteAsJson(MessageExist{false})
	} else {
		resp.WriteAsJson(MessageExist{true})
	}
}

func (h *handler) CreateNode(req *restful.Request, resp *restful.Response) {
	node := new(LinstorNode)
	err := req.ReadEntity(&node)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}
	client, ctx := linstorv1alpha1.GetClient(h.ControllerIP)
	err = linstorv1alpha1.CreateNode(ctx, client, node.Name, node.IP, node.NodeType)
	if err != nil {
		resp.WriteAsJson(err)
	}
}

func (h *handler) DeleteNode(req *restful.Request, resp *restful.Response) {
	nodename := req.PathParameter("node")
	fmt.Println(nodename)
	client, ctx := linstorv1alpha1.GetClient(h.ControllerIP)
	err := linstorv1alpha1.DeleteNode(ctx, client, string(nodename))
	if err != nil {
		resp.WriteAsJson(err)
	}
}

func (h *handler) handleListStorgePools(req *restful.Request, resp *restful.Response) {
	query := query.ParseQueryParameter(req)
	client, ctx := linstorv1alpha1.GetClient(h.ControllerIP)
	data := linstorv1alpha1.GetSPData(ctx, client)
	message := linstorv1alpha1.LinstorGetter{0, len(data), data}
	message.List(query)
	resp.WriteAsJson(message)
}

func (h *handler) DescribeStoragePool(req *restful.Request, resp *restful.Response) {
	storagepoolName := req.PathParameter("storagepool")
	client, ctx := linstorv1alpha1.GetClient(h.ControllerIP)
	exist := linstorv1alpha1.DescribeStoragePool(ctx, client, storagepoolName)
	resp.WriteAsJson(MessageExist{exist})
}

func (h *handler) GetAvailableStoragePools(req *restful.Request, resp *restful.Response) {
	reNodename := new(DiskfulSP)
	err := req.ReadEntity(&reNodename)
	client, ctx := linstorv1alpha1.GetClient(h.ControllerIP)
	allSPs := linstorv1alpha1.GetSPData(ctx, client)
	diskfulResources := linstorv1alpha1.GetResourcesDiskful(ctx, client)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	for _, sp := range allSPs {
		if sp["node"] == reNodename.NodeName {
			isDiskful := false
			for _, res := range diskfulResources {
				if res["storagepool"] == sp["name"] {
					isDiskful = true
					break
				}
			}
			if !isDiskful {
				data := sp
				resp.WriteAsJson(data)
			}
		}
	}

}

//func (h *handler) CreateStoragePool(req *restful.Request, resp *restful.Response) {
//	storagePool := new(LinstorSP)
//	err := req.ReadEntity(&storagePool)
//	if err != nil {
//		api.HandleBadRequest(resp, req, err)
//		return
//	}
//	client, ctx := linstorv1alpha1.GetClient(h.ControllerIP)
//	err = linstorv1alpha1.CreateSP(ctx, client, storagePool.Name, storagePool.NodeName, storagePool.Type, storagePool.Volume)
//	if err != nil {
//		resp.WriteAsJson(err)
//	}

func (h *handler) CreateStoragePool(req *restful.Request, resp *restful.Response) {
	storagePool := new(LinstorSP)
	err := req.ReadEntity(&storagePool)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}
	client, ctx := linstorv1alpha1.GetClient(h.ControllerIP)
	Name := storagePool.Metadata["name"].(string)
	err = linstorv1alpha1.CreateSP(ctx, client, Name, storagePool.NodeName, storagePool.Type, storagePool.Volume)
	if err != nil {
		resp.WriteAsJson(err)
	}
}
func (h *handler) DeleteStoragePool(req *restful.Request, resp *restful.Response) {
	nodeName := req.PathParameter("node")
	spName := req.PathParameter("storagepool")
	client, ctx := linstorv1alpha1.GetClient(h.ControllerIP)
	err := linstorv1alpha1.DeleteSP(ctx, client, spName, nodeName)
	if err != nil {
		resp.WriteAsJson(err)
	}
}

//func (h *handler) UpdateNode(req *restful.Request, resp *restful.Response) {
//	id := req.PathParameter("node")
//	fmt.Println(id)
//
//	linstor_node := new(LinstorNode)
//	err := req.ReadEntity(&linstor_node)
//
//	fmt.Println("-----")
//	fmt.Println(linstor_node)
//	if err != nil {
//		api.HandleBadRequest(resp, req, err)
//		return
//	}
//	// 执行
//	fmt.Println("------")
//
//
//}

func (h *handler) handleListResources(req *restful.Request, resp *restful.Response) {
	query := query.ParseQueryParameter(req)
	client, ctx := linstorv1alpha1.GetClient(h.ControllerIP)
	data := linstorv1alpha1.GetResources(ctx, client)
	message := linstorv1alpha1.LinstorGetter{0, len(data), data}
	message.List(query)
	resp.WriteAsJson(message)
}

func (h *handler) handleListResourcesDiskful(req *restful.Request, resp *restful.Response) {
	query := query.ParseQueryParameter(req)
	client, ctx := linstorv1alpha1.GetClient(h.ControllerIP)
	data := linstorv1alpha1.GetResourcesDiskful(ctx, client)
	message := linstorv1alpha1.LinstorGetter{0, len(data), data}
	message.List(query)
	resp.WriteAsJson(message)
}

func (h *handler) handleListResourcesDiskless(req *restful.Request, resp *restful.Response) {
	query := query.ParseQueryParameter(req)
	client, ctx := linstorv1alpha1.GetClient(h.ControllerIP)
	data := linstorv1alpha1.GetResourceDiskless(ctx, client)
	message := linstorv1alpha1.LinstorGetter{0, len(data), data}
	message.List(query)
	resp.WriteAsJson(message)
}

func (h *handler) DescribeResource(req *restful.Request, resp *restful.Response) {
	resname := req.PathParameter("resource")
	client, ctx := linstorv1alpha1.GetClient(h.ControllerIP)
	err := linstorv1alpha1.DescribeResource(ctx, client, resname)
	if err != nil {
		resp.WriteAsJson(MessageExist{false})
	} else {
		resp.WriteAsJson(MessageExist{true})
	}
}

func (h *handler) CreateResource(req *restful.Request, resp *restful.Response) {
	res := new(LinstorRes)
	err := req.ReadEntity(&res)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	client, ctx := linstorv1alpha1.GetClient(h.ControllerIP)
	Name := res.Metadata["name"].(string)
	err = linstorv1alpha1.CreateResource(ctx, client, Name, res.StoragePool, res.Size)
	if err != nil {
		resp.WriteAsJson(err)
		return
	}
	//if res.Node != nil {
	//	for _, node := range res.Node {
	//		err = linstorv1alpha1.CreateDisklessResource(ctx, client, res.Name, node)
	//	}
	//}
	//if err != nil {
	//	resp.WriteAsJson(err)
	//	return
	//}
	fmt.Println("linstor audit run....")
	lnau := auditing.GetLinstorAudit()
	isenable := lnau.Enabled()
	fmt.Println("isenable: ", isenable)

	resp.WriteEntity(servererr.None)
}

func (h *handler) CreateDiskless(req *restful.Request, resp *restful.Response) {
	res := new(LinstorRes)
	err := req.ReadEntity(&res)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}
	client, ctx := linstorv1alpha1.GetClient(h.ControllerIP)
	Name := res.Metadata["name"].(string)
	for _, node := range res.Node {
		err = linstorv1alpha1.CreateDisklessResource(ctx, client, Name, node)
		if err != nil {
			resp.WriteAsJson(err)
		} else {
			resp.WriteAsJson(node)
		}
	}
}

func (h *handler) IncreaseReplicas(req *restful.Request, resp *restful.Response) {
	res := new(ReplicaRes)
	err := req.ReadEntity(&res)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	client, ctx := linstorv1alpha1.GetClient(h.ControllerIP)
	err = linstorv1alpha1.UpdateDiskfulResource(ctx, client, res.ResName, res.NodeName, res.PoolName, res.TargetReplicas,
		res.CurrentReplicas)
	if err != nil {
		resp.WriteAsJson(err)
		return
	}
}

func (h *handler) DeleteResource(req *restful.Request, resp *restful.Response) {
	//nodeName := req.PathParameter("node")
	resName := req.PathParameter("resource")
	client, ctx := linstorv1alpha1.GetClient(h.ControllerIP)
	err := linstorv1alpha1.DeleteResource(ctx, client, resName)
	if err != nil {
		resp.WriteAsJson(err)
	}
}

func (h *handler) handleListLvmDevices(req *restful.Request, resp *restful.Response) {
	query := query.ParseQueryParameter(req)
	client, ctx := linstorv1alpha1.GetClient(h.ControllerIP)
	data := linstorv1alpha1.GetLvmDevices(ctx, client)
	message := linstorv1alpha1.LinstorGetter{0, len(data), data}
	message.List(query)
	resp.WriteAsJson(message)
}

func (h *handler) handleListLvmPVs(req *restful.Request, resp *restful.Response) {
	query := query.ParseQueryParameter(req)
	client, ctx := linstorv1alpha1.GetClient(h.ControllerIP)
	data := linstorv1alpha1.GetLvmPVs(ctx, client)
	message := linstorv1alpha1.LinstorGetter{0, len(data), data}
	message.List(query)
	resp.WriteAsJson(message)
}

func (h *handler) handleListLvmVGs(req *restful.Request, resp *restful.Response) {
	query := query.ParseQueryParameter(req)
	client, ctx := linstorv1alpha1.GetClient(h.ControllerIP)
	data := linstorv1alpha1.GetLvmVGs(ctx, client)
	message := linstorv1alpha1.LinstorGetter{0, len(data), data}
	message.List(query)
	resp.WriteAsJson(message)
}

func (h *handler) handleListLvmLVs(req *restful.Request, resp *restful.Response) {

	query := query.ParseQueryParameter(req)
	client, ctx := linstorv1alpha1.GetClient(h.ControllerIP)
	data := linstorv1alpha1.GetLvmLVs(ctx, client)
	linstorv1alpha1.CreatePV(ctx, client, "/dev/sdh", "ben2")
	message := linstorv1alpha1.LinstorGetter{0, len(data), data}
	message.List(query)
	resp.WriteAsJson(message)
}

func (h *handler) CreateResourceLvmPV(req *restful.Request, resp *restful.Response) {
	pv := new(LvmPV)
	err := req.ReadEntity(&pv)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	client, ctx := linstorv1alpha1.GetClient(h.ControllerIP)
	err = linstorv1alpha1.CreatePV(ctx, client, pv.Name, pv.Node)

	if err != nil {
		resp.WriteAsJson(err)
		return
	}

}

func (h *handler) CreateResourceLvmVG(req *restful.Request, resp *restful.Response) {
	vg := new(LvmVG)
	err := req.ReadEntity(&vg)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	client, ctx := linstorv1alpha1.GetClient(h.ControllerIP)
	err = linstorv1alpha1.CreateVG(ctx, client, vg.PV, vg.Name, vg.Node)

	if err != nil {
		resp.WriteAsJson(err)
		return
	}

}

func (h *handler) CreateResourceLvmThinPool(req *restful.Request, resp *restful.Response) {
	thinpool := new(LvmThinPool)
	err := req.ReadEntity(&thinpool)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	client, ctx := linstorv1alpha1.GetClient(h.ControllerIP)
	err = linstorv1alpha1.CreateThinPool(ctx, client, thinpool.Size, thinpool.Name, thinpool.VG, thinpool.Node)

	if err != nil {
		resp.WriteAsJson(err)
		return
	}

}

func (h *handler) CreateResourceLvmLV(req *restful.Request, resp *restful.Response) {
	lv := new(LvmLV)
	err := req.ReadEntity(&lv)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	client, ctx := linstorv1alpha1.GetClient(h.ControllerIP)
	err = linstorv1alpha1.CreateLV(ctx, client, lv.Size, lv.Name, lv.ThinPool, lv.VG, lv.Node)

	if err != nil {
		resp.WriteAsJson(err)
		return
	}

}
