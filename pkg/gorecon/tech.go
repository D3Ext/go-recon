package gorecon

import (
	"github.com/D3Ext/go-recon/core"
)

func GetTech(url string, timeout int) (map[string]struct{}, error) {
	return core.GetTech(url, timeout)
}
