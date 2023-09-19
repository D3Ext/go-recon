package gorecon

import (
	"github.com/D3Ext/go-recon/core"
)

func FindSecrets(url string, timeout int) ([]string, error) {
	return core.FindSecrets(url, timeout)
}
