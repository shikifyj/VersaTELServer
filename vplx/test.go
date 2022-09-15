package main

import (
	gp "kubesphere.io/kubesphere/pkg/utils/gopythonutil"
	"github.com/DataDog/go-python3"
	"fmt"
   	//"encoding/json"
)

//type Message struct {
//	Code int `json:"code"`
//	Count int `json:"count"`
//	Data []map[string]interface{} `json:"data"`
//}

func init(){
	fmt.Println("init")
	gp.Initialize()
	gp.ImportSystemModule()
	gp.ImportCustomModule("/home/samba/kubesphere.io/kubesphere/vplx")
	//gp.ImportCustomModule("/home/samba/kubesphere.io/kubesphere/pythoncode")
}


type LinstorNode struct {
	NodeName string `json:"node_name"`
	IP string `json:"ip"`
	NodeType string `json:"node_type"`
}



func CreateNode() {
	node := new(LinstorNode)
	node.NodeName = "ubuntu"
	node.IP = "10.203.1.158"
	node.NodeType = "Combined"

	//if err != nil {
	//	api.HandleBadRequest(resp, req, err)
	//	return
	//}

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

	nodeName := python3.PyUnicode_FromString(node.NodeName)
	nodeIP := python3.PyUnicode_FromString(node.IP)
	nodeType := python3.PyUnicode_FromString(node.NodeType)

	Data := classNode.CallMethodArgs("create_node",nodeObj,nodeName,nodeIP,nodeType)
	Result := python3.PyDict_GetItem(Data,python3.PyUnicode_FromString("result"))
	Info := python3.PyDict_GetItem(Data,python3.PyUnicode_FromString("info"))

	fmt.Println(python3.PyUnicode_AsUTF8(Result))
	fmt.Println(python3.PyUnicode_AsUTF8(Info))

	//resp.WriteAsJson(message)
}

func main() {
	CreateNode()
}
