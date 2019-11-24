package main

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var router *gin.Engine

func init() {
	gin.SetMode(gin.TestMode)
	os.Setenv("HTTP_PROXY", "http://localhost:3128")
	os.Setenv("PROXY_APPS", "a:b:c d:e:f")
	router = setupRouter()
}

func performRequest(router http.Handler, method, path string, body io.Reader) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, body)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	return w
}

func TestHealthCheck(t *testing.T) {
	w := performRequest(router, "GET", "/hc", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

func TestGetVersion(t *testing.T) {
	w := performRequest(router, "GET", "/version", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "version")
}

func TestReverseProxy(t *testing.T) {
	w := performRequest(router, "GET", "/v2/", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}
