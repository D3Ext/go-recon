package gorecon

import (
	"github.com/D3Ext/go-recon/core"
)

func FilterUrls(urls []string, filters []string) []string {
	return core.FilterUrls(urls, filters)
}
