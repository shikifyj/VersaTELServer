/*
Copyright 2019 The KubeSphere Authors.

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

package iputil

import (
	"fmt"
	"net"
	"net/http"
)

const (
	XForwardedFor = "X-Forwarded-For"
	XRealIP       = "X-Real-IP"
	XClientIP     = "x-client-ip"
)

func RemoteIp(req *http.Request) string {
	remoteAddr := req.RemoteAddr

	if ip := req.Header.Get(XClientIP); ip != "" {
		remoteAddr = ip
	} else if ip := req.Header.Get(XRealIP); ip != "" {
		remoteAddr = ip
	} else if ip = req.Header.Get(XForwardedFor); ip != "" {
		remoteAddr = ip
	} else {
		remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
	}

	if remoteAddr == "::1" {
		remoteAddr = "127.0.0.1"
	}

	return remoteAddr
}

func RemoteLoginIp(req *http.Request) string {

	fmt.Println("is remote ip!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")

	remoteAddr := req.RemoteAddr
	fmt.Println("remoteAddr: ", remoteAddr)
	remoteAddr, _, _ = net.SplitHostPort(remoteAddr)

	fmt.Println("XClientIP: ", req.Header.Get(XClientIP))
	fmt.Println("XRealIP: ", req.Header.Get(XRealIP))
	fmt.Println("XForwardedFor: ", req.Header.Get(XForwardedFor))

	return req.Header.Get(XRealIP)
}
