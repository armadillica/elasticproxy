package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"time"

	stdlog "log"

	log "github.com/sirupsen/logrus"
)

const elasticProxyVersion = "1.0"

var cliArgs struct {
	version    bool
	verbose    bool
	debug      bool
	elasticURL string
	listen     string
}

func parseCliArgs() {
	flag.BoolVar(&cliArgs.version, "version", false, "Shows the application version, then exits.")
	flag.BoolVar(&cliArgs.verbose, "verbose", false, "Enable info-level logging.")
	flag.BoolVar(&cliArgs.debug, "debug", false, "Enable debug-level logging.")
	flag.StringVar(&cliArgs.elasticURL, "elastic", "http://elastic:9200/", "URL of the ElasticSearch instance to proxy.")
	flag.StringVar(&cliArgs.listen, "listen", "[::]:9200", "Address to listen on.")
	flag.Parse()
}

func configLogging() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	// Only log the warning severity or above by default.
	level := log.WarnLevel
	if cliArgs.debug {
		level = log.DebugLevel
	} else if cliArgs.verbose {
		level = log.InfoLevel
	}
	log.SetLevel(level)
	stdlog.SetOutput(log.StandardLogger().Writer())
}

func logStartup(elasticURL *url.URL) {
	level := log.GetLevel()
	defer log.SetLevel(level)

	log.SetLevel(log.InfoLevel)
	log.WithFields(log.Fields{
		"elastic": elasticURL.String(),
		"version": elasticProxyVersion,
	}).Info("Starting ElasticProxy")
}

func main() {
	parseCliArgs()
	if cliArgs.version {
		fmt.Println(elasticProxyVersion)
		return
	}

	configLogging()

	elasticURL, err := url.Parse(cliArgs.elasticURL)
	if err != nil {
		log.Fatalf("Invalid URL %q: %s", cliArgs.elasticURL, err)
	}
	logStartup(elasticURL)

	// Set some more or less sensible limits & timeouts.
	http.DefaultTransport = &http.Transport{
		MaxIdleConns:          100,
		TLSHandshakeTimeout:   3 * time.Second,
		IdleConnTimeout:       15 * time.Minute,
		ResponseHeaderTimeout: 15 * time.Second,
	}

	logFields := log.Fields{"listen": cliArgs.listen}
	proxy := CreateElasticProxy(elasticURL)
	log.WithFields(logFields).Info("Starting HTTP server")
	err = http.ListenAndServe(cliArgs.listen, proxy)
	log.WithFields(logFields).WithError(err).Fatal("HTTP server failed")
}
