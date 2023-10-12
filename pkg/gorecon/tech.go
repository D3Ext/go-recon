package gorecon

import (
	"github.com/D3Ext/go-recon/core"
	"net/http"
)

// this function send a request to given url and returns running technologies
// Example: techs, err := GetTech("http://github.com", gorecon.DefaultClient())
func GetTech(url string, client *http.Client) (map[string]struct{}, error) {
	return core.GetTech(url, client)
}
