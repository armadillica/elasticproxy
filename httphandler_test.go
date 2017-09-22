package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	"github.com/stretchr/testify/assert"

	check "gopkg.in/check.v1"
	"gopkg.in/jarcoal/httpmock.v1"
)

type ElasticProxyTestSuite struct {
	proxy *ElasticProxy
}

var _ = check.Suite(&ElasticProxyTestSuite{})

func (s *ElasticProxyTestSuite) SetUpTest(c *check.C) {
	httpmock.Activate()

	url, err := url.Parse("http://elastic:9200/")
	assert.Nil(c, err)
	s.proxy = CreateElasticProxy(url)
}

func (s *ElasticProxyTestSuite) TearDownTest(c *check.C) {
	httpmock.DeactivateAndReset()
}

func (s *ElasticProxyTestSuite) TestGETHappy(t *check.C) {
	var respJSON struct {
		Field    string `json:"field"`
		Document struct {
			Subfield time.Time `json:"subfield"`
		} `json:"document"`
	}
	responder, err := httpmock.NewJsonResponder(200, respJSON)
	assert.Nil(t, err)
	httpmock.RegisterResponder(
		"GET", "http://elastic:9200/tasks/index/ABC123",
		responder,
	)

	respRec := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/tasks/index/ABC123", nil)
	s.proxy.ServeHTTP(respRec, request)

	assert.Equal(t, 200, respRec.Code)
	assert.Equal(t, "application/json", respRec.Header().Get("Content-Type"))
}
