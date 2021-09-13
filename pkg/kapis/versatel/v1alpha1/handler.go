package v1alpha1

import (
	"encoding/json"
	"fmt"
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
	NodeName string
	IP string
	NodeType string
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



func (h *handler) GetVersaTELURL (req *restful.Request, resp *restful.Response) {
	gp.Initialize()
	gp.ImportSystemModule()
	gp.ImportCustomModule("/home/samba/kubesphere.io/kubesphere/vplx")
	gp.ImportCustomModule("/home/samba/kubesphere.io/kubesphere/pythoncode")
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

	gp.Initialize()
	gp.ImportSystemModule()
	gp.ImportCustomModule("/home/samba/kubesphere.io/kubesphere/vplx")
	gp.ImportCustomModule("/home/samba/kubesphere.io/kubesphere/pythoncode")
	process := gp.GetModule("scheduler")
	classProcess := process.GetAttrString("VersaTEL")
	//// 实例化类
	if classProcess == nil {
		panic("could not retrieve 'ProcessData'")
	}
	defer classProcess.DecRef()
	EmptyTuple := gp.GetEmptyTuple()
	obj := classProcess.CallObject(EmptyTuple)
	if obj == nil {
		panic("could not retrieve 'obj'")
	}
	defer obj.DecRef()
	defer EmptyTuple.DecRef()
	Data := classProcess.CallMethodArgs("process_data_node",obj)
	Result := python3.PyUnicode_AsUTF8(Data)
	Message := linstorv1alpha1.LinstorGetter{}
	//Message := MessageList{}
	json.Unmarshal([]byte(Result),&Message)
	Message.List(query)
	resp.WriteAsJson(Message)
	fmt.Println(Message)

}

func (h *handler) CreateNode(req *restful.Request, resp *restful.Response) {
	node := new(LinstorNode)
	err := req.ReadEntity(&node)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	stor := gp.GetModule("execute.stor")
	if stor == nil {
		panic("could not retrieve 'execute.stor'")
	}
	defer stor.DecRef()


	classNode := stor.GetAttrString("Node")
	if classNode == nil {
		panic("could not retrieve 'Node'")
	}
	defer classNode.DecRef()

	EmptyTuple := gp.GetEmptyTuple()
	nodeObj := classNode.CallObject(EmptyTuple)
	if nodeObj == nil {
		panic("could not retrieve 'nodeObj'")
	}
	defer nodeObj.DecRef()
	defer EmptyTuple.DecRef()


	Data := classNode.CallMethodArgs("create_node",nodeObj,PyStr(node.NodeName),PyStr(node.IP),PyStr(node.NodeType))
	Result := python3.PyDict_GetItem(Data,PyStr("result"))
	Info := python3.PyDict_GetItem(Data,PyStr("info"))

	message := MessageOP{GoStr(Result),GoStr(Info)}
	resp.WriteAsJson(message)
}

func (h *handler) DeleteNode(req *restful.Request, resp *restful.Response) {
	nodename := req.PathParameter("node")

	stor := gp.GetModule("execute.stor")
	if stor == nil {
		panic("could not retrieve 'execute.stor'")
	}
	defer stor.DecRef()


	classNode := stor.GetAttrString("Node")
	if classNode == nil {
		panic("could not retrieve 'Node'")
	}
	defer classNode.DecRef()

	EmptyTuple := gp.GetEmptyTuple()
	nodeObj := classNode.CallObject(EmptyTuple)
	if nodeObj == nil {
		panic("could not retrieve 'nodeObj'")
	}
	defer nodeObj.DecRef()
	defer EmptyTuple.DecRef()

	Data := classNode.CallMethodArgs("delete_node",nodeObj,PyStr(nodename))
	Result := python3.PyDict_GetItem(Data,PyStr("result"))
	Info := python3.PyDict_GetItem(Data,PyStr("info"))

	message := MessageOP{GoStr(Result),GoStr(Info)}
	resp.WriteAsJson(message)
}

func (h *handler) UpdateNode(req *restful.Request, resp *restful.Response) {
	id := req.PathParameter("node")
	fmt.Println(id)

	linstor_node := new(LinstorNode)
	err := req.ReadEntity(&linstor_node)

	fmt.Println("-----")
	fmt.Println(linstor_node)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}
	// 执行
	fmt.Println("------")


}
