package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Token struct {
	Value string `json:"token"`
}

type App struct {
	Registry, Namespace string
}

const (
	KEY = "123456"
)

var (
	version          = "dev"
	commit           = "none"
	date             = "unknown"
	hostname, gitlab string
	token            Token
	apps             = make(map[string]App)
)

func setupRouter() *gin.Engine {
	entries := strings.Split(os.Getenv("PROXY_APPS"), " ")
	if len(entries[0]) < 1 {
		glog.Fatal("please set env.PROXY_APPS in format <app>:<registry>:<namespace> ...")
	}
	for _, e := range entries {
		parts := strings.Split(e, ":")
		apps[parts[0]] = App{parts[1], parts[2]}
	}
	r := gin.Default()
	r.NoRoute(gin.WrapF(reverseProxy))
	r.GET("/hc", healthCheck)
	r.GET("/version", getVersion)
	r.GET("/", statusOk)
	r.GET("/v2/", statusOk)
	return r
}

func main() {
	var port string
	flag.StringVar(&port, "port", "5001", "server listening port")
	flag.Parse()

	hostname, _ = os.Hostname()
	gitlab = base64.StdEncoding.EncodeToString(
		[]byte(os.Getenv("PROXY_GITLAB")))

	router := setupRouter()
	router.Run(":" + port)
}

func reverseProxy(writer http.ResponseWriter, request *http.Request) {
	var registry *url.URL
	keyIndex := 2
	elements := strings.Split(request.URL.Path, "/")
	if len(elements) > 4 {
		key := elements[keyIndex]
		if !validateKey(key) {
			http.Error(writer, "authorization required", http.StatusUnauthorized)
			return
		}
		if v, ok := apps[elements[keyIndex+1]]; ok {
			elements[keyIndex] = v.Namespace
			registry, _ = url.Parse("https://" + v.Registry)
			request.URL.Path = strings.Join(elements, "/")
		} else {
			http.Error(writer, "application not found", http.StatusNotFound)
			return
		}
	}

	request.URL.Scheme = registry.Scheme
	request.URL.Host = registry.Host
	request.Host = registry.Host
	request.Header.Set("X-Forwarded-Host", request.Header.Get("Host"))

	var resp *http.Response
	var req *http.Request
	var err error
	client := &http.Client{}

	glog.Info("--> GET " + request.URL.String())
	req, err = http.NewRequest("GET", request.URL.String(), nil)
	if token.Value != "" {
		req.Header.Set("Authorization", "Bearer "+token.Value)
	}

	resp, err = client.Do(req)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	glog.Infof("<-- %d", resp.StatusCode)

	// 401 -> auth server request
	if resp.StatusCode == http.StatusUnauthorized {
		url := parseChallenge(resp.Header.Get("Www-Authenticate"))
		glog.Info("--> auth " + url.String())
		// unless cache supports gitlab auth
		if url.Host == "gitlab.com" {
			c := &http.Client{Transport: &http.Transport{Proxy: nil}}
			r, _ := http.NewRequest("GET", url.String(), nil)
			glog.Infof("--> token %d", len(gitlab))
			if len(gitlab) > 0 {
				r.Header.Set("Authorization", "Basic "+gitlab)
			}
			resp, err = c.Do(r)
		} else {
			resp, err = http.Get(url.String())
		}
		if err != nil {
			http.Error(writer, err.Error(), http.StatusServiceUnavailable)
			return
		}
		glog.Infof("<-- auth %d", resp.StatusCode)

		if resp.StatusCode == http.StatusOK {
			json.NewDecoder(resp.Body).Decode(&token)
			req.Header.Set("Authorization", "Bearer "+token.Value)
			glog.Info("--> seq " + req.URL.String())
			resp, err = client.Do(req)
			glog.Infof("<-- seq %d", resp.StatusCode)
		} else {
			s, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				glog.Fatal(err)
			}
			http.Error(writer, string(s), resp.StatusCode)
			return
		}
	}

	for k, vv := range resp.Header {
		for _, v := range vv {
			writer.Header().Add(k, v)
		}
	}

	writer.WriteHeader(resp.StatusCode)
	io.Copy(writer, resp.Body)
}

func validateKey(t string) bool {
	sample := KEY
	return t == sample
}

func parseChallenge(s string) *url.URL {
	address := &url.URL{}
	values := url.Values{}
	entries := strings.Split(s, ",")
	for _, e := range entries {
		parts := strings.Split(e, "=")
		v, _ := strconv.Unquote(parts[1])
		if len(parts[0]) >= 6 && parts[0][:6] == "Bearer" {
			address, _ = url.Parse(v)
		} else {
			values.Add(parts[0], v)
		}
	}
	address.RawQuery = values.Encode()
	return address
}

func healthCheck(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}

func getVersion(c *gin.Context) {
	body := gin.H{
		"version":  version,
		"commit":   commit,
		"date":     date,
		"hostname": hostname,
		"ginmode":  gin.Mode(),
		"lang":     "golang",
	}
	c.JSON(http.StatusOK, body)
}

func statusOk(c *gin.Context) {
	c.Status(http.StatusOK)
}
