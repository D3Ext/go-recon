package gorecon

import (
	"github.com/D3Ext/go-recon/core"
	"net/http"
)

// main function to extract JS endpoints from a list of urls
// it receives a custom client for further customization
// Example: go gorecon.GetEndpointsFromFile(urls, results, 15, gorecon.DefaultClient())
func GetEndpoints(urls []string, results chan string, workers int, client *http.Client) error {
	return core.GetEndpoints(urls, results, workers, client)
}

// this function receives urls from channel so
// it's better for concurrency and configuration
func FetchEndpoints(urls <-chan string, results chan string, client *http.Client) error {
	return core.FetchEndpoints(urls, results, client)
}
