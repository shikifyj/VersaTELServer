package v1alpha1

import (
	"net/http"
	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"kubesphere.io/kubesphere/pkg/apiserver/query"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
)

const (
	GroupName = "storsecu.kubesphere.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha1"}

func AddToContainer(container *restful.Container, client kubernetes.Interface, config *rest.Config) error {
	webservice := runtime.NewWebService(GroupVersion)

	handler := newHandler(client, config)

	//tagsLinstor := []string{"linstor"}
	tagsLinstor := []string{"Clustered Resource"}

	webservice.Route(webservice.GET("/backup").
		To(handler.handleListbackupres).

		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor).
		Doc("Backup test.").
		Param(webservice.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
		Param(webservice.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(webservice.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(webservice.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
		//Param(webservice.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Returns(http.StatusOK, api.StatusOK, MessageList{}))


	webservice.Route(webservice.POST("/backup").
		To(handler.Backup).
		Doc("Create Backup.").
		Returns(http.StatusOK, api.StatusOK, MessageOP{}).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor).
		Reads(BackupInfo{}))

	webservice.Route(webservice.POST("/antivirus").
		To(handler.Backup).
		Doc("Create antivirus.").
		Returns(http.StatusOK, api.StatusOK, MessageOP{}).
		Metadata(restfulspec.KeyOpenAPITags, tagsLinstor).
		Reads(BackupInfo{}))


	container.Add(webservice)

	return nil
}



