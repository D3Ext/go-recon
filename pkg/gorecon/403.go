package gorecon

import (
  "net/http"
	"github.com/D3Ext/go-recon/core"
)

// try different ways to bypass 403 status code urls
// returns a slice of urls with payloads on them,
// a slice with their respective status codes, and
// finally an error
func Check403(url, word string, client *http.Client, user_agent string) ([]string, []int, error) {
	return core.Check403(url, word, client, user_agent)
}
