package gorecon

import (
	"github.com/D3Ext/go-recon/core"
	"net/http"
)

func CreateHttpClient(timeout int) *http.Client {
	return core.CreateHttpClient(timeout)
}

func CreateHttpClientWithProxy(timeout int, proxy string) (*http.Client, error) {
	return core.CreateHttpClientWithProxy(timeout, proxy)
}
