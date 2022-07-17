package http

import (
	"io"
	"log"
	"net"
	"net/http"

	core "github.com/bzEq/bx/core"
)

// See https://www.rfc-editor.org/rfc/rfc9110.html#field.connection
var HopByHopFields = []string{
	"Connection",
	"Proxy-Connection",
	"Keep-Alive",
	"TE",
	"Transfer-Encoding",
	"Upgrade",
}

func RemoveHopByHopFields(header http.Header) {
	for _, f := range HopByHopFields {
		header.Del(f)
	}
}

type HTTPProxy struct {
	Transport http.RoundTripper
	Dial      func(string, string) (net.Conn, error)
}

func (self *HTTPProxy) handleConnect(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	h, ok := w.(http.Hijacker)
	if !ok {
		log.Println("Hijacking not supported")
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	c, _, err := h.Hijack()
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	if self.Dial == nil {
		self.Dial = net.Dial
	}
	remoteConn, err := self.Dial("tcp", req.Host)
	if err != nil {
		log.Println(err)
		return
	}
	defer remoteConn.Close()
	core.RunSimpleProtocolSwitch(c, remoteConn, nil, nil)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func (self *HTTPProxy) handleNormal(w http.ResponseWriter, req *http.Request) {
	req.RequestURI = ""
	RemoveHopByHopFields(req.Header)
	client := &http.Client{Transport: self.Transport}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	RemoveHopByHopFields(resp.Header)
	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// Modified from
// https://www.sobyte.net/post/2021-09/https-proxy-in-golang-in-less-than-100-lines-of-code/
func (self *HTTPProxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodConnect {
		self.handleConnect(w, req)
	} else {
		self.handleNormal(w, req)
	}
}
