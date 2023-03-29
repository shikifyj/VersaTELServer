/*
Copyright 2020 KubeSphere Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package auditing

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
	v1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/apis/audit"
	"k8s.io/klog"

	devopsv1alpha3 "kubesphere.io/api/devops/v1alpha3"

	auditv1alpha1 "kubesphere.io/kubesphere/pkg/apiserver/auditing/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/client/listers/auditing/v1alpha1"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/devops"
	options "kubesphere.io/kubesphere/pkg/simple/client/auditing"
	"kubesphere.io/kubesphere/pkg/utils/iputil"
)

const (
	DefaultWebhook       = "kube-auditing-webhook"
	DefaultCacheCapacity = 10000
	CacheTimeout         = time.Second
)

type Auditing interface {
	Enabled() bool
	K8sAuditingEnabled() bool
	LogRequestObject(req *http.Request, info *request.RequestInfo) *auditv1alpha1.Event
	LogRequestObjectLinstor(req *http.Request, info *request.RequestInfo)
	LogResponseObject(e *auditv1alpha1.Event, resp *ResponseCapture)
}

type auditing struct {
	webhookLister v1alpha1.WebhookLister
	devopsGetter  v1alpha3.Interface
	cache         chan *auditv1alpha1.Event
	backend       *Backend
}

type BenEvent struct {
	// Devops project
	Name string
	// The workspace which this audit event happened
	Number string
	// The cluster which this audit event happened
	Age string
	// Message send to user.

}

type Installer struct {
    Spec     Spe `json:"spec"`
}
 
type Spe struct {

    Audit   AuditingLinstor `json:"auditing"`

}

type AuditingLinstor struct {

    Enable   bool `json:"enabled"`

}
var loginevent *auditv1alpha1.Event

var loginUserName string

var linstorAudit *auditing

var loginerr bool = true

var islogin bool = false

var loginip string

func NewAuditing(informers informers.InformerFactory, opts *options.Options, stopCh <-chan struct{}) Auditing {

	linstorAudit = &auditing{
		webhookLister: informers.KubeSphereSharedInformerFactory().Auditing().V1alpha1().Webhooks().Lister(),
		devopsGetter:  devops.New(informers.KubeSphereSharedInformerFactory()),
		cache:         make(chan *auditv1alpha1.Event, DefaultCacheCapacity),
	}

	linstorAudit.backend = NewBackend(opts, linstorAudit.cache, stopCh)
	return linstorAudit
}


// func NewAuditing(informers informers.InformerFactory, opts *options.Options, stopCh <-chan struct{}) Auditing {

// 	a := &auditing{
// 		webhookLister: informers.KubeSphereSharedInformerFactory().Auditing().V1alpha1().Webhooks().Lister(),
// 		devopsGetter:  devops.New(informers.KubeSphereSharedInformerFactory()),
// 		cache:         make(chan *auditv1alpha1.Event, DefaultCacheCapacity),
// 	}

// 	a.backend = NewBackend(opts, a.cache, stopCh)
// 	return a
// }


func GetLinstorAudit() Auditing{
	return linstorAudit
}

func SetLoginUserName(name string, err bool) {
	loginerr = err
	loginUserName = name
}

func SendLogout() {
	fmt.Println("is SendLogout")
	
	if islogin == true {
		islogin = false
		loginevent.SourceIPs[0] = loginip

		loginevent.ObjectRef.Name = loginUserName
		loginevent.User.Username = loginUserName

		loginevent.ResponseStatus.Reason = "Logout successfully"
		loginevent.RequestReceivedTimestamp = metav1.NowMicro()
		linstorAudit.cacheEvent(*loginevent)

	}

}



func (a *auditing) getAuditLevel() audit.Level {
	wh, err := a.webhookLister.Get(DefaultWebhook)
	if err != nil {
		klog.V(8).Info(err)
		return audit.LevelNone
	}

	return (audit.Level)(wh.Spec.AuditLevel)
}

func (a *auditing) Enabled() bool {

	level := a.getAuditLevel()
	return !level.Less(audit.LevelMetadata)
}

func (a *auditing) K8sAuditingEnabled() bool {
	wh, err := a.webhookLister.Get(DefaultWebhook)
	if err != nil {
		klog.V(8).Info(err)
		return false
	}

	return wh.Spec.K8sAuditingEnabled
}

// If the request is not a standard request, or a resource request,
// or part of the audit information cannot be obtained through url,
// the function that handles the request can obtain Event from
// the context of the request, assign value to audit information,
// including name, verb, resource, subresource, message etc like this.
//
//	info, ok := request.AuditEventFrom(request.Request.Context())
//	if ok {
//		info.Verb = "post"
//		info.Name = created.Name
//	}
//
func (a *auditing) LogRequestObjectLinstor(req *http.Request, info *request.RequestInfo) {



	ips := make([]string, 1)
	ips[0] = iputil.RemoteIp(req)

	user, ok := request.UserFrom(req.Context())


	insbody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		klog.Error(err)
	}
	_ = req.Body.Close()
	req.Body = ioutil.NopCloser(bytes.NewBuffer(insbody))


    var inst Installer
    if err = json.Unmarshal(insbody, &inst); err != nil {
        fmt.Printf("Unmarshal err, %v\n", err)
    }
    fmt.Printf("%+v", inst)

    if inst.Spec.Audit.Enable {


		e2 := &auditv1alpha1.Event{
			Devops:    info.DevOps,
			Workspace: info.Workspace,
			Cluster:   info.Cluster,
			Event: audit.Event{
				RequestURI:               info.Path,
				Verb:                     "",
				Level:                    a.getAuditLevel(),
				AuditID:                  types.UID(uuid.New().String()),
				Stage:                    audit.StageResponseComplete,
				ImpersonatedUser:         nil,
				UserAgent:                req.UserAgent(),
				RequestReceivedTimestamp: metav1.NowMicro(),
				Annotations:              nil,
				ObjectRef: &audit.ObjectReference{
					Resource:        "SystemFunction",
					Namespace:       info.Namespace,
					Name:            "audit",
					UID:             "",
					APIGroup:        info.APIGroup,
					APIVersion:      info.APIVersion,
					ResourceVersion: info.ResourceScope,
					Subresource:     info.Subresource,
				},
			},
		}


		if ok {
			e2.User.Username = user.GetName()
			e2.User.UID = user.GetUID()
			e2.User.Groups = user.GetGroups()
		}

		e2.SourceIPs = ips
		e2.ResponseStatus = &metav1.Status{Code: int32(0)}
		e2.ResponseStatus.Status = "INFO"
		e2.ResponseStatus.Reason = "Enable audit successfully"
		a.cacheEvent(*e2)

    }


}









func (a *auditing) LogRequestObject(req *http.Request, info *request.RequestInfo) *auditv1alpha1.Event {

	if !a.Enabled() && info.Path != "/apis/installer.kubesphere.io/v1alpha1/namespaces/kubesphere-system/clusterconfigurations/ks-installer" {
		return nil
	}

	// Ignore the dryRun k8s request.
	if info.IsKubernetesRequest {
		if len(req.URL.Query()["dryRun"]) != 0 {
			klog.V(6).Infof("ignore dryRun request %s", req.URL.Path)
			return nil
		}
	}

	e := &auditv1alpha1.Event{
		Devops:    info.DevOps,
		Workspace: info.Workspace,
		Cluster:   info.Cluster,
		Event: audit.Event{
			RequestURI:               info.Path,
			Verb:                     info.Verb,
			Level:                    a.getAuditLevel(),
			AuditID:                  types.UID(uuid.New().String()),
			Stage:                    audit.StageResponseComplete,
			ImpersonatedUser:         nil,
			UserAgent:                req.UserAgent(),
			RequestReceivedTimestamp: metav1.NowMicro(),
			Annotations:              nil,
			ObjectRef: &audit.ObjectReference{
				Resource:        info.Resource,
				Namespace:       info.Namespace,
				Name:            info.Name,
				UID:             "",
				APIGroup:        info.APIGroup,
				APIVersion:      info.APIVersion,
				ResourceVersion: info.ResourceScope,
				Subresource:     info.Subresource,
			},
		},
	}

	if info.Path == "/oauth/token" || info.Path == "/oauth/logout"{
		fmt.Println("is oauth after event: ", info.Path)
		fmt.Println("info.Resource, name: ", info.Resource, info.Name)

		tokenbody, err := ioutil.ReadAll(req.Body)
		if err != nil {
			fmt.Println("readall err!")
			klog.Error(err)
			return e
		}
		_ = req.Body.Close()
		req.Body = ioutil.NopCloser(bytes.NewBuffer(tokenbody))
		fmt.Println("body: %s", tokenbody)

	}

	// Get the workspace which the devops project be in.
	if len(e.Devops) > 0 && len(e.Workspace) == 0 {
		res, err := a.devopsGetter.List("", query.New())
		if err != nil {
			klog.Error(err)
		}

		for _, obj := range res.Items {
			d := obj.(*devopsv1alpha3.DevOpsProject)

			if d.Name == e.Devops {
				e.Workspace = d.Labels["kubesphere.io/workspace"]
			} else if d.Status.AdminNamespace == e.Devops {
				e.Workspace = d.Labels["kubesphere.io/workspace"]
				e.Devops = d.Name
			}
		}
	}

	ips := make([]string, 1)
	ips[0] = iputil.RemoteIp(req)
	e.SourceIPs = ips

	user, ok := request.UserFrom(req.Context())
	if ok {

		if info.Path == "/oauth/token" || info.Path == "/oauth/logout"{
			fmt.Println("user is ok, username: ", user.GetName())
		}
		e.User.Username = user.GetName()
		e.User.UID = user.GetUID()
		e.User.Groups = user.GetGroups()

		e.User.Extra = make(map[string]v1.ExtraValue)
		for k, v := range user.GetExtra() {
			e.User.Extra[k] = v
		}
	}
	if info.Verb == "update" {

		if info.Path == "/apis/installer.kubesphere.io/v1alpha1/namespaces/kubesphere-system/clusterconfigurations/ks-installer" {
			insbody, err := ioutil.ReadAll(req.Body)
			if err != nil {
				klog.Error(err)
				return e
			}
			_ = req.Body.Close()
			req.Body = ioutil.NopCloser(bytes.NewBuffer(insbody))


		    var inst Installer
		    if err = json.Unmarshal(insbody, &inst); err != nil {
		        fmt.Printf("Unmarshal err, %v\n", err)
		    }
		    fmt.Printf("%+v", inst)

		    if a.Enabled() !=  inst.Spec.Audit.Enable {


				e2 := &auditv1alpha1.Event{
					Devops:    info.DevOps,
					Workspace: info.Workspace,
					Cluster:   info.Cluster,
					Event: audit.Event{
						RequestURI:               info.Path,
						Verb:                     "",
						Level:                    a.getAuditLevel(),
						AuditID:                  types.UID(uuid.New().String()),
						Stage:                    audit.StageResponseComplete,
						ImpersonatedUser:         nil,
						UserAgent:                req.UserAgent(),
						RequestReceivedTimestamp: metav1.NowMicro(),
						Annotations:              nil,
						ObjectRef: &audit.ObjectReference{
							Resource:        "SystemFunction",
							Namespace:       info.Namespace,
							Name:            "audit",
							UID:             "",
							APIGroup:        info.APIGroup,
							APIVersion:      info.APIVersion,
							ResourceVersion: info.ResourceScope,
							Subresource:     info.Subresource,
						},
					},
				}


				if ok {
					e2.User.Username = user.GetName()
					e2.User.UID = user.GetUID()
					e2.User.Groups = user.GetGroups()
				}

				e2.SourceIPs = ips
				if inst.Spec.Audit.Enable{
					e2.Verb = "enable"
				}
				e2.ResponseStatus = &metav1.Status{Code: int32(0)}

				e2.ResponseStatus.Status = "INFO"
				e2.ResponseStatus.Reason = "Disable audit successfully"				
				a.cacheEvent(*e2)

		    }
        }
    }


	if info.Path == "/oauth/token" {
		fmt.Println("is /oauth/token!!!")



		loginevent = &auditv1alpha1.Event{
			Devops:    info.DevOps,
			Workspace: info.Workspace,
			Cluster:   info.Cluster,
			Event: audit.Event{
				RequestURI:               info.Path,
				Verb:                     "",
				Level:                    a.getAuditLevel(),
				AuditID:                  types.UID(uuid.New().String()),
				Stage:                    audit.StageResponseComplete,
				ImpersonatedUser:         nil,
				UserAgent:                req.UserAgent(),
				RequestReceivedTimestamp: metav1.NowMicro(),
				Annotations:              nil,
				ObjectRef: &audit.ObjectReference{
					Resource:        "users",
					Namespace:       info.Namespace,
					Name:            "",
					UID:             "",
					APIGroup:        info.APIGroup,
					APIVersion:      info.APIVersion,
					ResourceVersion: info.ResourceScope,
					Subresource:     info.Subresource,
				},
			},
		}


		if ok {
			loginevent.User.Username = "waitforchange"
			loginevent.User.UID = user.GetUID()
			loginevent.User.Groups = user.GetGroups()
		}
		ips[0] = getloginip()
		loginevent.SourceIPs = ips

		loginevent.ResponseStatus = &metav1.Status{Code: int32(0)}

		loginevent.ResponseStatus.Status = "INFO"
		loginevent.ResponseStatus.Reason = "Login successfully"
		return loginevent
	}




	if info.Path == "/oauth/logout" {
		fmt.Println("is /oauth/logout for event!")

		


		e4 := &auditv1alpha1.Event{
			Devops:    info.DevOps,
			Workspace: info.Workspace,
			Cluster:   info.Cluster,
			Event: audit.Event{
				RequestURI:               info.Path,
				Verb:                     "",
				Level:                    a.getAuditLevel(),
				AuditID:                  types.UID(uuid.New().String()),
				Stage:                    audit.StageResponseComplete,
				ImpersonatedUser:         nil,
				UserAgent:                req.UserAgent(),
				RequestReceivedTimestamp: metav1.NowMicro(),
				Annotations:              nil,
				ObjectRef: &audit.ObjectReference{
					Resource:        "users",
					Namespace:       info.Namespace,
					Name:            "",
					UID:             "",
					APIGroup:        info.APIGroup,
					APIVersion:      info.APIVersion,
					ResourceVersion: info.ResourceScope,
					Subresource:     info.Subresource,
				},
			},
		}


		if ok {
			e4.User.Username = user.GetName()
			e4.ObjectRef.Name = user.GetName()
			e4.User.UID = user.GetUID()
			e4.User.Groups = user.GetGroups()
		}

		e4.SourceIPs = ips
		e4.SourceIPs[0] = loginip

		e4.ResponseStatus = &metav1.Status{Code: int32(0)}

		e4.ResponseStatus.Status = "INFO"
		e4.ResponseStatus.Reason = "Logout successfully"
		islogin = false
		a.cacheEvent(*e4)

	}


	if (e.Level.GreaterOrEqual(audit.LevelRequest) || e.Verb == "create") && req.ContentLength > 0 {
		body, err := ioutil.ReadAll(req.Body)
	
		if err != nil {
			klog.Error(err)
			return e
		}
		_ = req.Body.Close()
		req.Body = ioutil.NopCloser(bytes.NewBuffer(body))

		if e.Level.GreaterOrEqual(audit.LevelRequest) {
			e.RequestObject = &runtime.Unknown{Raw: body}
		}

		// For resource creating request, get resource name from the request body.
		if info.Verb == "create" {
			obj := &auditv1alpha1.Object{}
			if err := json.Unmarshal(body, obj); err == nil {
				e.ObjectRef.Name = obj.Name
			}
		}
	}

	if info.Path == "/oauth/token" || info.Path == "/oauth/logout"{
		fmt.Println("end for token or logout", info.Path)
	}

	return e
}

func (a *auditing) LogResponseObject(e *auditv1alpha1.Event, resp *ResponseCapture) {

	e.StageTimestamp = metav1.NowMicro()


	if e.User.Username == "waitforchange" {
		if islogin != true {
			fmt.Println("is not islogin")

			if loginerr {
				e.ResponseStatus.Reason = "Login failed"
			}else {
				islogin = true
				loginip = e.SourceIPs[0]
			}
			
			e.ObjectRef.Name = loginUserName
			e.User.Username = loginUserName
			a.cacheEvent(*e)
		}

	} else {

		e.ResponseStatus = &metav1.Status{Code: int32(resp.StatusCode())}
		//e.ResponseStatus.reason = e.Verb + " " + resp.StatusCode()
		if e.Level.GreaterOrEqual(audit.LevelRequestResponse) {
			e.ResponseObject = &runtime.Unknown{Raw: resp.Bytes()}
		}

		a.cacheEvent(*e)
	}
}

func (a *auditing) cacheEvent(e auditv1alpha1.Event) {

	select {
	case a.cache <- &e:
		return
	case <-time.After(CacheTimeout):
		klog.V(8).Infof("cache audit event %s timeout", e.AuditID)
		break
	}
}

type ResponseCapture struct {
	http.ResponseWriter
	wroteHeader bool
	status      int
	body        *bytes.Buffer
}

func NewResponseCapture(w http.ResponseWriter) *ResponseCapture {
	return &ResponseCapture{
		ResponseWriter: w,
		wroteHeader:    false,
		body:           new(bytes.Buffer),
	}
}

func (c *ResponseCapture) Header() http.Header {
	return c.ResponseWriter.Header()
}

func (c *ResponseCapture) Write(data []byte) (int, error) {

	c.WriteHeader(http.StatusOK)
	c.body.Write(data)
	return c.ResponseWriter.Write(data)
}

func (c *ResponseCapture) WriteHeader(statusCode int) {
	if !c.wroteHeader {
		c.status = statusCode
		c.ResponseWriter.WriteHeader(statusCode)
		c.wroteHeader = true
	}
}

func (c *ResponseCapture) Bytes() []byte {
	return c.body.Bytes()
}

func (c *ResponseCapture) StatusCode() int {
	return c.status
}

// Hijack implements the http.Hijacker interface.  This expands
// the Response to fulfill http.Hijacker if the underlying
// http.ResponseWriter supports it.
func (c *ResponseCapture) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := c.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("ResponseWriter doesn't support Hijacker interface")
	}
	return hijacker.Hijack()
}

// CloseNotify is part of http.CloseNotifier interface
func (c *ResponseCapture) CloseNotify() <-chan bool {
	//nolint:staticcheck
	return c.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

func getloginip() string{
	ip := ""
    resp, err := http.Get("http://getloginip.default.svc:80")
    if err != nil {
        // 处理错误
        fmt.Println("getloginip request err: ", err)
        return ip
    }
    defer resp.Body.Close()

    // 读取响应内容
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        // 处理错误
        fmt.Println("read resp err: ", err)
        return ip     
    }
    fmt.Println(string(body))
    return string(body)
    
}