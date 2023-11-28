package proxy

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Server struct {
	c      *Config
	p      *Proxy
	server *http.Server
	wg     sync.WaitGroup
}

func NewServer(c *Config) (svr *Server) {
	if c.ServerPort == "" {
		c.ServerPort = ":2233"
	}
	if !strings.HasPrefix(c.ServerPort, ":") {
		c.ServerPort = ":" + c.ServerPort
	}
	if c.ProxySchema == "" {
		c.ProxySchema = SchemaHTTP
	}
	if c.ProxyPort == "" {
		c.ProxyPort = ":80"
	}
	if !strings.HasPrefix(c.ProxyPort, ":") {
		c.ProxyPort = ":" + c.ProxyPort
	}

	p := &Proxy{
		httpCli: &http.Client{
			Timeout: 60 * time.Second,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: defaultTransportDialContext(&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}),
				TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				DisableKeepAlives:     true,
				ForceAttemptHTTP2:     true,
			},
		},
		Schema:       c.ProxySchema,
		Host:         c.ProxyHost,
		Port:         c.ProxyPort,
		ProxyAuthKey: c.ProxyAuthKey,
		ShowLog:      c.ShowLog,
		log:          log.New(os.Stdout, string([]byte{27, 91, 51, 53, 109})+" [PROXY] "+string([]byte{27, 91, 48, 109}), log.Ldate|log.Lmicroseconds),
	}

	svr = &Server{
		c: c,
		p: p,
		server: &http.Server{
			Addr:         c.ServerPort,
			Handler:      p,
			ReadTimeout:  time.Minute,
			WriteTimeout: time.Minute,
		},
		wg: sync.WaitGroup{},
	}
	return
}

func (s *Server) ListenAndServe() {
	http.Handle("/", s.p)
	// monitoring signal
	go s.goNotifySignal()
	// start gin http server
	log.Printf("Listening and serving HTTP on %s", s.c.ServerPort)
	if err := s.server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			panic(fmt.Sprintf("server.ListenAndServe(), error(%+v).", err))
		}
		log.Println("http: Server closed")
	}
	log.Println("wait for process working finished")
	// wait for process finished
	s.wg.Wait()
	log.Println("process exit")
}

func (s *Server) Close() {
	if s.server != nil {
		// disable keep-alives on existing connections
		s.server.SetKeepAlivesEnabled(false)
		_ = s.server.Shutdown(context.Background())
	}
}
