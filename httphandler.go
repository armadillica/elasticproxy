package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	log "github.com/sirupsen/logrus"
)

// ElasticProxy forwards received HTTP calls to another HTTP server.
type ElasticProxy struct {
	url   *url.URL
	proxy *httputil.ReverseProxy
}

// CreateElasticProxy returns a new elasticProxy object.
func CreateElasticProxy(elasticURL *url.URL) *ElasticProxy {
	return &ElasticProxy{
		url:   elasticURL,
		proxy: httputil.NewSingleHostReverseProxy(elasticURL),
	}
}

// ServeHTTP only forwards allowed requests to the real ElasticSearch server.
func (ep *ElasticProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	status := 0

	startTime := time.Now().UTC()
	defer func() {
		endTime := time.Now().UTC()
		duration := endTime.Sub(startTime)
		if status == 0 {
			log.Infof("%s %s %s (proxied) %v", r.RemoteAddr, r.Method, r.URL.String(), duration)
		} else {
			log.Infof("%s %s %s %d %v", r.RemoteAddr, r.Method, r.URL.String(), status, duration)
		}
	}()

	if r.Method != "GET" {
		status = http.StatusMethodNotAllowed
		w.WriteHeader(status)
		return
	}

	if r.Header.Get("Upgrade") == "websocket" {
		log.Warningf("%s request from %s with Upgrade: websocket", r.Method, r.RemoteAddr)
		status = http.StatusNotImplemented
		w.WriteHeader(status)
		fmt.Fprintln(w, "Websockets not supported")
		return
	}

	// All our checks were fine, so now we can defer to the ReverseProxy to do the actual work.
	r.Header.Set("Host", ep.url.Host)
	ep.proxy.ServeHTTP(w, r)
}
