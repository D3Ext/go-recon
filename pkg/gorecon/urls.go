package gorecon

import (
	"github.com/D3Ext/go-recon/core"
	"github.com/D3Ext/go-recon/core/providers"
	"net/http"
)

// main function to enumerate urls about provided domain, urls are sent through channel
// set "recursive" to false if you don't want to get urls related to subdomains
func GetAllUrls(domain string, results chan string, client *http.Client, recursive bool) error {
	return core.GetAllUrls(domain, results, client, recursive)
}

/*

providers functions

*/

func WaybackUrls(domain string, results chan string, client *http.Client, workers int, recursive bool) error {
	return providers.WaybackUrls(domain, results, client, workers, recursive)
}

func AlienVaultUrls(domain string, results chan string, client *http.Client, recursive bool) error {
	return providers.AlienVaultUrls(domain, results, client, recursive)
}

func UrlScanUrls(domain string, results chan string, client *http.Client, recursive bool, apikey string) error {
	return providers.UrlScanUrls(domain, results, client, recursive, apikey)
}
