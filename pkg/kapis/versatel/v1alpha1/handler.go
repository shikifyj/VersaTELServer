package v1alpha1

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	linstorv1alpha1 "kubesphere.io/kubesphere/pkg/models/versatel/v1alpha1/linstor"
)

//var PyStr = python3.PyUnicode_FromString
//var PyInt = python3.PyLong_FromLong
//var GoStr = python3.PyUnicode_AsUTF8
//var GoInt = python3.PyLong_AsLong

type handler struct {
	//linstorGetterV1alpha1  *linstorv1alpha1.linstorGetter
}


func newHandler() handler{
	return handler{}
}

type MessageList struct {
	Code int `json:"code"`
	Count int `json:"count"`
	Data []map[string]string `json:"data"`
}


type MessageOP struct {
	Result string `json:"result"`
	Info string `json:"info"`
}

type MessageExist struct {
	Exist bool `json:"exist"`
}

type LinstorNode struct {
	Name string `json:"name"`
	IP string `json:"ip"`
	NodeType string `json:"node_type"`
}

type LinstorSP struct {
	Name string `json:"name"`
	NodeName string `json:"node"`
	Type string `json:"type"`
	Volume string `json:"volume"`
}

type LinstorRes struct {
	Name string `json:"name"`
	//Node string `json:"node"`
	StoragePool []string `json:"storagepool"`
	Size string `json:"size"`
}

type URLResponse struct {
	URL string `json:"URL"`
}

//func init(){
//	gp.Initialize()
//	gp.ImportSystemModule()
//	gp.ImportCustomModule("/home/samba/kubesphere.io/kubesphere/vplx")
//	gp.ImportCustomModule("/home/samba/kubesphere.io/kubesphere/pythoncode")
//}



func (h *handler) handleListNodes (req *restful.Request, resp *restful.Response) {
	query := query.ParseQueryParameter(req)
	client, ctx := linstorv1alpha1.GetClient()
	data := linstorv1alpha1.GetNodeData(ctx,client)
	message := linstorv1alpha1.LinstorGetter{0,len(data),data}
	message.List(query)
	resp.WriteAsJson(message)
}

func (h *handler) DescribeNode(req *restful.Request, resp *restful.Response) {
	nodename := req.PathParameter("node")
	client, ctx := linstorv1alpha1.GetClient()
	err := linstorv1alpha1.DescribeNode(ctx,client,nodename)
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
	client, ctx := linstorv1alpha1.GetClient()
	err = linstorv1alpha1.CreateNode(ctx,client,node.Name,node.IP,node.NodeType)
	if err != nil{
		resp.WriteAsJson(err)
	}
}

func (h *handler) DeleteNode(req *restful.Request, resp *restful.Response) {
	nodename := req.PathParameter("node")
	fmt.Println(nodename)
	client, ctx := linstorv1alpha1.GetClient()
	err := linstorv1alpha1.DeleteNode(ctx,client,string(nodename))
	if err != nil {
		resp.WriteAsJson(err)
	}
}



func (h *handler) handleListStorgePools (req *restful.Request, resp *restful.Response) {
	query := query.ParseQueryParameter(req)
	client, ctx := linstorv1alpha1.GetClient()
	data := linstorv1alpha1.GetSPData(ctx,client)
	message := linstorv1alpha1.LinstorGetter{0,len(data),data}
	message.List(query)
	resp.WriteAsJson(message)
}

func (h *handler) DescribeStoragePool(req *restful.Request, resp *restful.Response) {
	storagepoolName := req.PathParameter("storagepool")
	client, ctx := linstorv1alpha1.GetClient()
	exist := linstorv1alpha1.DescribeStoragePool(ctx,client,storagepoolName)
	resp.WriteAsJson(MessageExist{exist})
}

func (h *handler) CreateStoragePool(req *restful.Request, resp *restful.Response) {
	storagePool := new(LinstorSP)
	err := req.ReadEntity(&storagePool)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}
	client, ctx := linstorv1alpha1.GetClient()
	err = linstorv1alpha1.CreateSP(ctx,client,storagePool.Name,storagePool.NodeName,storagePool.Type,storagePool.Volume)
	if err != nil{
		resp.WriteAsJson(err)
	}
}



func (h *handler) DeleteStoragePool(req *restful.Request, resp *restful.Response) {
	nodeName := req.PathParameter("node")
	spName := req.PathParameter("storagepool")
	client, ctx := linstorv1alpha1.GetClient()
	err := linstorv1alpha1.DeleteSP(ctx,client,spName,nodeName)
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


func (h *handler) handleListResources (req *restful.Request, resp *restful.Response) {
	query := query.ParseQueryParameter(req)
	client, ctx := linstorv1alpha1.GetClient()
	data := linstorv1alpha1.GetResources(ctx,client)
	message := linstorv1alpha1.LinstorGetter{0,len(data),data}
	message.List(query)
	resp.WriteAsJson(message)
}

func (h *handler) DescribeResource(req *restful.Request, resp *restful.Response) {
	resname := req.PathParameter("resource")
	client, ctx := linstorv1alpha1.GetClient()
	err := linstorv1alpha1.DescribeResource(ctx,client,resname)
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
	client, ctx := linstorv1alpha1.GetClient()
	err = linstorv1alpha1.CreateResource(ctx,client,res.Name,res.StoragePool,res.Size)
	if err != nil{
		resp.WriteAsJson(err)
	}
}


func (h *handler) DeleteResource(req *restful.Request, resp *restful.Response) {
	//nodeName := req.PathParameter("node")
	resName := req.PathParameter("resource")
	client, ctx := linstorv1alpha1.GetClient()
	err := linstorv1alpha1.DeleteResource(ctx,client,resName)
	if err != nil {
		resp.WriteAsJson(err)
	}
}

