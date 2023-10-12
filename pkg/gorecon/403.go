package gorecon

import (
	"github.com/D3Ext/go-recon/core"
)

// try different ways to bypass 403 status code urls
// returns slice of urls with payloads on them,
// a slice with their respective status codes, and
// finally an error
func Check403(url, word string, timeout int) ([]string, []int, error) {
	return core.Check403(url, word, timeout)
}
