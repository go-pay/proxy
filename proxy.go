package proxy

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
)

type Proxy struct {
	httpCli      *http.Client
	Schema       SchemaType // SchemaHTTP or SchemaHTTPS
	Host         string
	Port         string
	ProxyAuthKey string
	ShowLog      bool
	log          *log.Logger
	mu           sync.Mutex
}

func (s *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Proxy(r.Context(), w, r)
}

func (s *Proxy) Proxy(c context.Context, w http.ResponseWriter, r *http.Request) {
	var (
		rMethod = r.Method
		rHeader = r.Header
		rUri    = r.RequestURI
	)
	// 验证 Proxy-Auth-Key
	authKey := rHeader.Get(HeaderAuthKey)
	if s.ProxyAuthKey != authKey {
		http.Error(w, fmt.Sprintf("[%s] Invalid Proxy-Auth-Key", authKey), http.StatusUnauthorized)
		return
	}
	host := string(s.Schema) + s.Host + s.Port
	uri := host + rUri
	// Request
	req, err := http.NewRequestWithContext(c, r.Method, uri, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Request Header
	req.Header = r.Header
	resp, err := s.httpCli.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	if s.ShowLog {
		s.log.Printf("| %d | %s |  Proxy to %s\n", resp.StatusCode, rMethod, host+r.RequestURI)
	}
	// Response Header
	for k, vs := range resp.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (s *Proxy) clientIP(req *http.Request) string {
	clientIP := req.Header.Get("X-Forwarded-For")
	clientIP = strings.TrimSpace(strings.Split(clientIP, ",")[0])
	if clientIP == "" {
		clientIP = strings.TrimSpace(req.Header.Get("X-Real-Ip"))
	}
	if clientIP != "" {
		return clientIP
	}
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(req.RemoteAddr)); err == nil {
		return ip
	}
	return ""
}
