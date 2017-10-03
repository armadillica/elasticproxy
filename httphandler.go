package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	log "github.com/sirupsen/logrus"
)

var allowedMethods = map[string]bool{
	"GET":  true,
	"HEAD": true,
}
var allowedPostPaths = map[string]bool{
	"/_mget":            true,
	"/_msearch":         true,
	"/.kibana/_search":  true,
	"/.kibana/_msearch": true,
}

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
	fields := log.Fields{
		"remote_addr": r.RemoteAddr,
		"method":      r.Method,
		"url":         r.URL.String(),
	}
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		fields["x_forwarded_for"] = xff
	}

	startTime := time.Now().UTC()

	defer func() {
		endTime := time.Now().UTC()
		fields["duration"] = endTime.Sub(startTime)

		if status == 0 {
			log.WithFields(fields).Info("Request proxied")
		} else {
			fields["status"] = status
			log.WithFields(fields).Warning("Request blocked")
		}
	}()

	if r.Method == "POST" && allowedPostPaths[r.URL.Path] {
		log.WithFields(fields).Info("Allowing POST request")
	} else if !allowedMethods[r.Method] {
		status = http.StatusMethodNotAllowed
		w.WriteHeader(status)
		return
	}

	if r.Header.Get("Upgrade") == "websocket" {
		log.WithFields(fields).Warning("Upgrade to websocket blocked")
		status = http.StatusNotImplemented
		w.WriteHeader(status)
		fmt.Fprintln(w, "Websockets not supported")
		return
	}

	// All our checks were fine, so now we can defer to the ReverseProxy to do the actual work.
	r.Header.Set("Host", ep.url.Host)
	ep.proxy.ServeHTTP(w, r)
}
