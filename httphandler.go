package main

import (
	"io"
	"net/http"
	"net/url"
	"time"

	log "github.com/sirupsen/logrus"
)

// ElasticProxy forwards received HTTP calls to another HTTP server.
type ElasticProxy struct {
	url *url.URL
}

// CreateElasticProxy returns a new elasticProxy object.
func CreateElasticProxy(elasticURL *url.URL) *ElasticProxy {
	return &ElasticProxy{url: elasticURL}
}

// ServeHTTP only forwards allowed requests to the real ElasticSearch server.
func (ep *ElasticProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	status := 0

	startTime := time.Now().UTC()
	defer func() {
		if status == 0 {
			status = http.StatusInternalServerError
			w.WriteHeader(status)
		}
		endTime := time.Now().UTC()
		duration := endTime.Sub(startTime)
		log.Infof("%s %s %s %d %v", r.RemoteAddr, r.Method, r.URL.String(), status, duration)
	}()

	if r.Method != "GET" {
		status = http.StatusMethodNotAllowed
		w.WriteHeader(status)
		return
	}

	if r.URL == nil {
		status = http.StatusBadRequest
		w.WriteHeader(status)
		log.Warningf("%s request from %s without URL received", r.Method, r.RemoteAddr)
		return
	}

	proxyURL := ep.url.ResolveReference(r.URL).String()

	// Perform request to the real ElasticSearch
	proxyReq, err := http.NewRequest(r.Method, proxyURL, r.Body)
	if err != nil {
		log.Warningf("unable to create %s request to %s: %s", r.Method, proxyURL, err)
		return
	}
	proxyReq.Header = r.Header
	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		log.Warningf("unable to perform %s request to %s: %s", r.Method, proxyURL, err)
		return
	}

	// TODO: add timeout check
	status = resp.StatusCode
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Warningf("unable to copy body from %s request to %s: %s", r.Method, proxyURL, err)
		return
	}
}
