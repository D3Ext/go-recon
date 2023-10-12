package gorecon

import (
	"github.com/D3Ext/go-recon/core"
	"net/http"
)

// this function checks if given url is vulnerable to open redirect with provided payloads
// if keyword has value, it will be replaced with payloads
// Example: vuln_urls, err := CheckRedirect("http://example.com/index.php?p=FUZZ", client, []string{"bing.com", "//bing.com"}, "FUZZ")
func CheckRedirect(url string, c *http.Client, payloads []string, keyword string) ([]string, error) {
	return core.CheckRedirect(url, c, payloads, keyword)
}

// return all defined payloads
func GetPayloads() []string {
	return core.GetPayloads()
}

// return common payloads
func GetCommonPayloads() []string {
	return core.GetCommonPayloads()
}
