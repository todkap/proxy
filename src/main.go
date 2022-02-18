package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
)

// These constant is used to define server
const (
	
	PORT    = "9081"
)

type DebugTransport struct{}

func (DebugTransport) RoundTrip(r *http.Request) (*http.Response, error) {
        b, err := httputil.DumpRequestOut(r, false)
        if err != nil {
                return nil, err
        }
        fmt.Println(string(b))
        return http.DefaultTransport.RoundTrip(r)
}

// Serve a reverse proxy for a given url
func serveReverseProxy(target string, res http.ResponseWriter, req *http.Request) {
	// parse the url

	url, _ := url.Parse(target)

	// create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ErrorHandler = ErrHandle

	if(getenvBool("PROXY_DEBUG")){
		proxy.Transport = DebugTransport{}
	}
	// NewSingleHostReverseProxy does not rewrite the Host header. 
	// To rewrite Host headers, use ReverseProxy directly with a 
	// custom Director policy.
	req.Host = req.URL.Host

	// Note that ServeHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(res, req)
}

func ErrHandle(res http.ResponseWriter, req *http.Request, err error) {
	fmt.Println(err)
}   

// Log the typeform payload and redirect url
func logRequestPayload(proxyURL string) {
	log.Printf("proxy_url: %s\n", proxyURL)
}

// Balance returns one of the servers based using round-robin algorithm
func getProxyURL() string {
	return os.Getenv("END_POINT")
}

func getenvBool(key string) (bool) {
    s, err := getenvStr(key)
    if err != nil {
        return false
    }
    v, err := strconv.ParseBool(s)
    if err != nil {
        return false
    }
    return v
}
var ErrEnvVarEmpty = errors.New("getenv: environment variable empty")

func getenvStr(key string) (string, error) {
    v := os.Getenv(key)
    if v == "" {
        return v, ErrEnvVarEmpty
    }
    return v, nil
}

// Given a request send it to the appropriate url
func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	url := getProxyURL()
	logRequestPayload(url)
	serveReverseProxy(url, res, req)
}

func main() {
	// start server
	http.HandleFunc("/", handleRequestAndRedirect)
	log.Fatal(http.ListenAndServe(":"+PORT, nil))
}