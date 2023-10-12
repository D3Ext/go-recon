package gorecon

import (
	"github.com/D3Ext/go-recon/core"
	"net/http"
)

// this function send a request to url with an LFI payload
// to try to trigger the possible WAF (Web Application Firewall) i.e. Cloudflare
// Example: waf, err := gorecon.DetectWaf(url, "", "", gorecon.DefaultHttpClient())
func DetectWaf(url string, payload string, keyword string, client *http.Client) (string, error) {
	return core.DetectWaf(url, payload, keyword, client)
}
