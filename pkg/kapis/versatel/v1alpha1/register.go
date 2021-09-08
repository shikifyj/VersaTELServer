package v1alpha1

import (
	"net/http"
	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/runtime/schema"
	restfulspec "github.com/emicklei/go-restful-openapi"

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

	webservice.Route(webservice.GET("/vtelurl").
		Reads("").
		To(handler.GetVersaTELURL).
		Returns(http.StatusOK, api.StatusOK, URLResponse{}).
		Doc("Api for versatel url"))

	webservice.Route(webservice.GET("/linstor/node").
		Reads("").
		To(handler.handleListNodes).
		Returns(http.StatusOK, api.StatusOK, MessageList{}).
		Doc("Get all linstor node"))

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




	//webservice.Route(webservice.PUT("/linstornode/{node}").
	//	To(handler.UpdateNode).
	//	Doc("Update node").
	//	Param(webservice.PathParameter("node", "linstor node name")).
	//	Returns(http.StatusOK, api.StatusOK, LinstorNode{}).
	//	Metadata(restfulspec.KeyOpenAPITags, tags))


	container.Add(webservice)

	return nil
}
