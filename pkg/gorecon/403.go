package gorecon

import (
	"github.com/D3Ext/go-recon/core"
)

func Check403(url, word string, timeout int) ([]string, []int, error) {
	return core.Check403(url, word, timeout)
}
