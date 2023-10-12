package gorecon

import (
	"github.com/D3Ext/go-recon/core"
	"net/http"
)

// this function receives a url and a client to look for
// potential leaked secrets like API keys (using regex)
// Example: secrets, err := gorecon.FindSecrets("http://github.com", gorecon.DefaultClient())
func FindSecrets(url string, client *http.Client) ([]string, error) {
	return core.FindSecrets(url, client)
}
