package v1alpha1

import (
	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"net/http"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
)

const (
	GroupName = "versatel.kubesphere.io"
)



var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha1"}

func AddToContainer(container *restful.Container) error {
	webservice := runtime.NewWebService(GroupVersion)
	handler := newHandler()

	tagsLinstor := []string{"linstor"}

	webservice.Route(webservice.GET("/linstor/node").
		To(handler.handleListNodes).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor).
		Doc("Cluster level resources").
		Param(webservice.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
		Param(webservice.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(webservice.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(webservice.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		//Param(webservice.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, api.StatusOK, MessageList{}))

	webservice.Route(webservice.GET("/linstor/node/{node}").
		To(handler.DescribeNode).
		Doc("Retrieve node details.").
		Param(webservice.PathParameter("node", "nodename")).
		Returns(http.StatusOK, api.StatusOK, MessageExist{}).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor))

	webservice.Route(webservice.POST("/linstor/node").
		To(handler.CreateNode).
		Doc("Create a linstor node.").
		Returns(http.StatusOK, api.StatusOK, MessageOP{}).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor).
		Reads(LinstorNode{}))

	webservice.Route(webservice.DELETE("/linstor/node/{node}").
		To(handler.DeleteNode).
		Doc("Delete the specified node.").
		Param(webservice.PathParameter("node", "nodename")).
		Returns(http.StatusOK, api.StatusOK, MessageOP{}).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor))


	webservice.Route(webservice.GET("linstor/storagepool").
		To(handler.handleListStorgePools).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor).
		Doc("Cluster level resources").
		Param(webservice.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
		Param(webservice.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(webservice.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(webservice.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		//Param(webservice.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, api.StatusOK, MessageList{}))

	webservice.Route(webservice.GET("/linstor/storagepool/{storagepool}").
		To(handler.DescribeStoragePool).
		Doc("Retrieve storagepool details.").
		Param(webservice.PathParameter("storagepool", "storagepoolname")).
		Returns(http.StatusOK, api.StatusOK, MessageExist{}).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor))


	webservice.Route(webservice.POST("/linstor/storagepool").
		To(handler.CreateStoragePool).
		Doc("Create a linstor storagepool.").
		Returns(http.StatusOK, api.StatusOK, MessageOP{}).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor).
		Reads(LinstorSP{}))

	webservice.Route(webservice.DELETE("/linstor/storagepool/{storagepool}/{node}").
		To(handler.DeleteStoragePool).
		Doc("Delete the specified storagepool.").
		Param(webservice.PathParameter("node", "nodename")).
		Param(webservice.PathParameter("storagepool", "storagepoolname")).
		Returns(http.StatusOK, api.StatusOK, MessageOP{}).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor))


	webservice.Route(webservice.GET("linstor/resource").
		To(handler.handleListResources).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor).
		Doc("Cluster level resources").
		Param(webservice.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
		Param(webservice.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(webservice.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(webservice.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		//Param(webservice.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, api.StatusOK, MessageList{}))

	webservice.Route(webservice.GET("/linstor/resource/{resource}").
		To(handler.DescribeResource).
		Doc("Retrieve resource details.").
		Param(webservice.PathParameter("resource", "resourcename")).
		Returns(http.StatusOK, api.StatusOK, MessageExist{}).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor))

	webservice.Route(webservice.POST("/linstor/resource").
		To(handler.CreateResource).
		Doc("Create a linstor resource.").
		Returns(http.StatusOK, api.StatusOK, MessageOP{}).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor).
		Reads(LinstorRes{}))

	webservice.Route(webservice.DELETE("/linstor/resource/{resource}").
		To(handler.DeleteResource).
		Doc("Delete the specified storagepool.").
		//Param(webservice.PathParameter("node", "nodename")).
		Param(webservice.PathParameter("resource", "resourcename")).
		Returns(http.StatusOK, api.StatusOK, MessageOP{}).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor))




	//webservice.Route(webservice.PUT("/linstornode/{node}").
	//	To(handler.UpdateNode).
	//	Doc("Update node").
	//	Param(webservice.PathParameter("node", "linstor node name")).
	//	Returns(http.StatusOK, api.StatusOK, LinstorNode{}).
	//	Metadata(restfulspec.KeyOpenAPITags, tags))


	container.Add(webservice)

	return nil
}
