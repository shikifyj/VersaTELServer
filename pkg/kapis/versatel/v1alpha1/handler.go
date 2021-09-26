package v1alpha1

import (
	"fmt"
	//"fmt"
	"github.com/DataDog/go-python3"
	"github.com/emicklei/go-restful"
	"kubesphere.io/kubesphere/pkg/api"
	gp "kubesphere.io/kubesphere/pkg/utils/gopythonutil"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	linstorv1alpha1 "kubesphere.io/kubesphere/pkg/models/versatel/v1alpha1/linstor"
)

var PyStr = python3.PyUnicode_FromString
var PyInt = python3.PyLong_FromLong
var GoStr = python3.PyUnicode_AsUTF8
var GoInt = python3.PyLong_AsLong

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

type LinstorNode struct {
	Name string `json:"name"`
	IP string `json:"ip"`
	NodeType string `json:"node_type"`
}

type LinstorSP struct {
	Name string `json:"name"`
	NodeName string `json:"node_name"`
	Kind string `json:"kind"`
	Volume string `json:"volume"`
}


type URLResponse struct {
	URL string `json:"URL"`
}

func init(){
	gp.Initialize()
	gp.ImportSystemModule()
	gp.ImportCustomModule("/home/samba/kubesphere.io/kubesphere/vplx")
	gp.ImportCustomModule("/home/samba/kubesphere.io/kubesphere/pythoncode")
}



func (h *handler) GetVersaTELURL (req *restful.Request, resp *restful.Response) {
	vtel := gp.GetModule("vtel")
	//vtel := gp.ImportModule("/home/samba/kubesphere.io/kubesphere/pythoncode", "vtel") // 导入Python文件，获取模块对象
	if vtel == nil {
		panic("could not retrieve 'vtel'")
	}
	defer vtel.DecRef()
	getMgtUrl := vtel.GetAttrString("get_vtelmgt_URL") // 获取模块的函数
	EmptyTuple := gp.GetEmptyTuple() // 生成一个空元组
	PyURL := getMgtUrl.Call(EmptyTuple,EmptyTuple) // 执行函数
	GoURL := python3.PyUnicode_AsUTF8(PyURL) // 转化成 Go 的类型数据
	resp.WriteAsJson(URLResponse{
		URL: GoURL,
	})
}

func (h *handler) handleListNodes (req *restful.Request, resp *restful.Response) {
	query := query.ParseQueryParameter(req)
	client, ctx := linstorv1alpha1.GetClient()
	data := linstorv1alpha1.GetNodeData(ctx,client)
	message := linstorv1alpha1.LinstorGetter{0,len(data),data}
	message.List(query)
	resp.WriteAsJson(message)
}

func (h *handler) CreateNode(req *restful.Request, resp *restful.Response) {
	node := new(LinstorNode)
	err := req.ReadEntity(&node)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}
	client, ctx := linstorv1alpha1.GetClient()
	errCreate := linstorv1alpha1.CreateNode(ctx,client,node.Name,node.IP,node.NodeType)
	if errCreate != nil{
		resp.WriteAsJson(err)
	}
}

//func (h *handler) DeleteNode(req *restful.Request, resp *restful.Response) {
//	nodename := req.PathParameter("node")
//
//	stor := gp.GetModule("execute.stor")
//	if stor == nil {
//		panic("could not retrieve 'execute.stor'")
//	}
//	defer stor.DecRef()
//
//
//	classNode := stor.GetAttrString("Node")
//	if classNode == nil {
//		panic("could not retrieve 'Node'")
//	}
//	defer classNode.DecRef()
//
//	EmptyTuple := gp.GetEmptyTuple()
//	nodeObj := classNode.CallObject(EmptyTuple)
//	if nodeObj == nil {
//		panic("could not retrieve 'nodeObj'")
//	}
//	defer nodeObj.DecRef()
//	defer EmptyTuple.DecRef()
//
//	Data := classNode.CallMethodArgs("delete_node",nodeObj,PyStr(nodename))
//	Result := python3.PyDict_GetItem(Data,PyStr("result"))
//	Info := python3.PyDict_GetItem(Data,PyStr("info"))
//
//	message := MessageOP{GoStr(Result),GoStr(Info)}
//	resp.WriteAsJson(message)
//}

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


func (h *handler) CreateStoragePool(req *restful.Request, resp *restful.Response) {
	storagePool := new(LinstorSP)
	err := req.ReadEntity(&storagePool)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	//client, ctx := linstorv1alpha1.GetClient()
	//errCreate := linstorv1alpha1.CreateSP(ctx,client,node.Name,node.IP,node.NodeType)
	//if errCreate != nil{
	//	resp.WriteAsJson(err)
	//}
}

//func (h *handler) DeleteNode(req *restful.Request, resp *restful.Response) {
//	nodename := req.PathParameter("node")
//
//	stor := gp.GetModule("execute.stor")
//	if stor == nil {
//		panic("could not retrieve 'execute.stor'")
//	}
//	defer stor.DecRef()
//
//
//	classNode := stor.GetAttrString("Node")
//	if classNode == nil {
//		panic("could not retrieve 'Node'")
//	}
//	defer classNode.DecRef()
//
//	EmptyTuple := gp.GetEmptyTuple()
//	nodeObj := classNode.CallObject(EmptyTuple)
//	if nodeObj == nil {
//		panic("could not retrieve 'nodeObj'")
//	}
//	defer nodeObj.DecRef()
//	defer EmptyTuple.DecRef()
//
//	Data := classNode.CallMethodArgs("delete_node",nodeObj,PyStr(nodename))
//	Result := python3.PyDict_GetItem(Data,PyStr("result"))
//	Info := python3.PyDict_GetItem(Data,PyStr("info"))
//
//	message := MessageOP{GoStr(Result),GoStr(Info)}
//	resp.WriteAsJson(message)
//}

func (h *handler) DeleteStoragePool(req *restful.Request, resp *restful.Response) {
	nodename := req.PathParameter("storagepool")
	fmt.Println(nodename)
	client, ctx := linstorv1alpha1.GetClient()
	err := linstorv1alpha1.DeleteNode(ctx,client,string(nodename))
	if err != nil {
		panic(err)
	}
	message := MessageOP{"SUCCESS",""}
	resp.WriteAsJson(message)
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