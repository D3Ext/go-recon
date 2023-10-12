package gorecon

import (
	"github.com/D3Ext/go-recon/core"
)

// remove useless urls, duplicates and more
// to optimize results as much as possible from
// a list of urls
// Example: new_urls := gorecon.FilterUrls(urls, []string{"hasparams"})
func FilterUrls(urls []string, filters []string) []string {
	return core.FilterUrls(urls, filters)
}
