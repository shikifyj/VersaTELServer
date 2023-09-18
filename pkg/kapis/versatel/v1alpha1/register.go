package v1alpha1

import (
	"net/http"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"kubesphere.io/kubesphere/pkg/apiserver/query"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
)

const (
	GroupName = "versatel.kubesphere.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha1"}

func GetLinstorIP() string {

	ip, err := ioutil.ReadFile("/etc/linstorip/linstorip")
	if err != nil {
		log.Fatal(err)
	}
	return string(ip)
}

func AddToContainer(container *restful.Container, ip string) error {
	webservice := runtime.NewWebService(GroupVersion)
	//linstorip := GetLinstorIP()
	//ip = linstorip
	handler := newHandler(ip)

	tagsLinstor := []string{"Clustered Resource"}

	webservice.Route(webservice.GET("/versasdsnode").
		To(handler.handleListNodes).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor).
		Doc("Cluster level resources").
		Param(webservice.PathParameter("versatel", "cluster level resource type, e.g. pods,jobs,configmaps,services.")).
		Param(webservice.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
		Param(webservice.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(webservice.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(webservice.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		Param(webservice.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, api.StatusOK, MessageList{}))

	webservice.Route(webservice.GET("/versasdsnode/{node}").
		To(handler.DescribeNode).
		Doc("Retrieve node details.").
		Param(webservice.PathParameter("node", "nodename")).
		Returns(http.StatusOK, api.StatusOK, MessageExist{}).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor))

	webservice.Route(webservice.POST("/versasdsnode").
		To(handler.CreateNode).
		Doc("Create a linstor node.").
		Returns(http.StatusOK, api.StatusOK, MessageOP{}).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor).
		Reads(LinstorNode{}))

	webservice.Route(webservice.DELETE("/versasdsnode/{node}").
		To(handler.DeleteNode).
		Doc("Delete the specified node.").
		Param(webservice.PathParameter("node", "nodename")).
		Returns(http.StatusOK, api.StatusOK, MessageOP{}).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor))

	webservice.Route(webservice.GET("versasdsstoragepool").
		To(handler.handleListStorgePools).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor).
		Doc("Cluster level resources").
		Param(webservice.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
		Param(webservice.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(webservice.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(webservice.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		//Param(webservice.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, api.StatusOK, MessageList{}))

	webservice.Route(webservice.GET("/versasdsstoragepool/{storagepool}").
		To(handler.DescribeStoragePool).
		Doc("Retrieve storagepool details.").
		Param(webservice.PathParameter("storagepool", "storagepoolname")).
		Returns(http.StatusOK, api.StatusOK, MessageExist{}).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor))

	webservice.Route(webservice.POST("/versasdsstoragepool").
		To(handler.CreateStoragePool).
		Doc("Create a linstor storagepool.").
		Returns(http.StatusOK, api.StatusOK, MessageOP{}).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor).
		Reads(LinstorSP{}))

	webservice.Route(webservice.POST("/versasdsstoragepool/nodiskful").
		To(handler.GetAvailableStoragePools).
		Doc("Retrieve no diskful resource storagepool details.").
		Returns(http.StatusOK, api.StatusOK, MessageOP{}).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor).
		Reads(DiskfulSP{}))

	webservice.Route(webservice.POST("/versasdsresource/copy").
		To(handler.IncreaseReplicas).
		Doc("Create a copy.").
		Returns(http.StatusOK, api.StatusOK, MessageOP{}).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor).
		Reads(ReplicaRes{}))

	webservice.Route(webservice.DELETE("/versasdsstoragepool/{storagepool}/{node}").
		To(handler.DeleteStoragePool).
		Doc("Delete the specified storagepool.").
		Param(webservice.PathParameter("node", "nodename")).
		Param(webservice.PathParameter("storagepool", "storagepoolname")).
		Returns(http.StatusOK, api.StatusOK, MessageOP{}).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor))

	webservice.Route(webservice.GET("versasdsresource").
		To(handler.handleListResources).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor).
		Doc("Cluster level resources").
		Param(webservice.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
		Param(webservice.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(webservice.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(webservice.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		//Param(webservice.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, api.StatusOK, MessageList{}))

	webservice.Route(webservice.GET("versasdsresource/diskful").
		To(handler.handleListResourcesDiskful).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor).
		Doc("Cluster level resources").
		Param(webservice.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
		Param(webservice.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(webservice.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(webservice.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		//Param(webservice.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, api.StatusOK, MessageList{}))

	webservice.Route(webservice.GET("versasdsresource/diskless").
		To(handler.handleListResourcesDiskless).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor).
		Doc("Cluster level resources").
		Param(webservice.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
		Param(webservice.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(webservice.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(webservice.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		//Param(webservice.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, api.StatusOK, MessageList{}))

	webservice.Route(webservice.GET("/versasdsresource/{resource}").
		To(handler.DescribeResource).
		Doc("Retrieve resource details.").
		Param(webservice.PathParameter("resource", "resourcename")).
		Returns(http.StatusOK, api.StatusOK, MessageExist{}).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor))

	webservice.Route(webservice.POST("/versasdsresource").
		To(handler.CreateResource).
		Doc("Create a linstor resource.").
		Returns(http.StatusOK, api.StatusOK, MessageOP{}).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor).
		Reads(LinstorRes{}))

	webservice.Route(webservice.POST("/versasdsresource/diskless").
		To(handler.CreateDiskless).
		Doc("Create a linstor diskless resource.").
		Returns(http.StatusOK, api.StatusOK, MessageOP{}).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor).
		Reads(LinstorRes{}))

	webservice.Route(webservice.DELETE("/versasdsresource/{resource}").
		To(handler.DeleteResource).
		Doc("Delete the specified storagepool.").
		//Param(webservice.PathParameter("node", "nodename")).
		Param(webservice.PathParameter("resource", "resourcename")).
		Returns(http.StatusOK, api.StatusOK, MessageOP{}).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor))

	webservice.Route(webservice.GET("/device").
		To(handler.handleListLvmDevices).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor).
		Doc("Cluster level lvm").
		Param(webservice.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
		Param(webservice.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(webservice.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(webservice.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		//Param(webservice.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, api.StatusOK, MessageList{}))

	webservice.Route(webservice.GET("/pv").
		To(handler.handleListLvmPVs).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor).
		Doc("Cluster level pv").
		Param(webservice.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
		Param(webservice.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(webservice.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(webservice.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		//Param(webservice.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, api.StatusOK, MessageList{}))

	webservice.Route(webservice.GET("/vg").
		To(handler.handleListLvmVGs).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor).
		Doc("Cluster level vg").
		Param(webservice.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
		Param(webservice.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(webservice.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(webservice.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		//Param(webservice.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, api.StatusOK, MessageList{}))

	webservice.Route(webservice.GET("/thinpool").
		To(handler.handleListLvmLVs).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor).
		Doc("Cluster level lv").
		Param(webservice.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
		Param(webservice.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(webservice.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(webservice.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		//Param(webservice.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, api.StatusOK, MessageList{}))

	webservice.Route(webservice.POST("/pv").
		To(handler.CreateResourceLvmPV).
		Doc("Create pvs.").
		Returns(http.StatusOK, api.StatusOK, MessageOP{}).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor).
		Reads(LvmPV{}))

	webservice.Route(webservice.POST("/vg").
		To(handler.CreateResourceLvmVG).
		Doc("Create pvs.").
		Returns(http.StatusOK, api.StatusOK, MessageOP{}).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor).
		Reads(LvmVG{}))

	webservice.Route(webservice.POST("/thinpool").
		To(handler.CreateResourceLvmThinPool).
		Doc("Create pvs.").
		Returns(http.StatusOK, api.StatusOK, MessageOP{}).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor).
		Reads(LvmThinPool{}))

	webservice.Route(webservice.POST("/lv").
		To(handler.CreateResourceLvmLV).
		Doc("Create pvs.").
		Returns(http.StatusOK, api.StatusOK, MessageOP{}).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor).
		Reads(LvmLV{}))
	//webservice.Route(webservice.PUT("/linstornode/{node}").
	//	To(handler.UpdateNode).
	//	Doc("Update node").
	//	Param(webservice.PathParameter("node", "linstor node name")).
	//	Returns(http.StatusOK, api.StatusOK, LinstorNode{}).
	//	Metadata(restfulspec.KeyOpenAPITags, tags))

	webservice.Route(webservice.DELETE("/pv/{node}/{name}").
		To(handler.DeletePV).
		Doc("Delete the PV.").
		Param(webservice.PathParameter("name", "pv_name")).
		Param(webservice.PathParameter("node", "node_name")).
		Returns(http.StatusOK, api.StatusOK, MessageOP{}).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor))

	webservice.Route(webservice.DELETE("/vg/{node}/{name}").
		To(handler.DeleteVG).
		Doc("Delete the VG.").
		Param(webservice.PathParameter("name", "vg_name")).
		Param(webservice.PathParameter("node", "node_name")).
		Returns(http.StatusOK, api.StatusOK, MessageOP{}).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor))

	webservice.Route(webservice.DELETE("/thinpool/{node}/{vg_name}/{name}").
		To(handler.DeleteThinPool).
		Doc("Delete the VG.").
		Param(webservice.PathParameter("name", "thinpool_name")).
		Param(webservice.PathParameter("node", "node_name")).
		Param(webservice.PathParameter("vg_name", "vg_name")).
		Returns(http.StatusOK, api.StatusOK, MessageOP{}).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor))

	container.Add(webservice)

	return nil
}
