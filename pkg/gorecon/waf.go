package gorecon

import (
	"github.com/D3Ext/go-recon/core"
)

func DetectWaf(url string, payload string, keyword string, timeout int) (string, error) {
	return core.DetectWaf(url, payload, keyword, timeout)
}
