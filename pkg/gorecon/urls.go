package gorecon

import (
  "github.com/D3Ext/go-recon/core"
)

// results := make(chan string)
// go gorecon.GetAllUrls("example.com", results, 5000, true)
func GetAllUrls(domain string, results chan string, timeout int, recursive bool) {
  core.GetAllUrls(domain, results, timeout, recursive)
}

// results := make(chan string)
// go gorecon.GetAllUrls("example.com", results, 5000, true)
func GetWaybackUrls(domain string, results chan string, timeout int, recursive bool) error {
  return core.GetWaybackUrls(domain, results, timeout, recursive)
}

func GetOTXUrls(domain string, timeout int, recursive bool) ([]string, error) {
  return core.GetOTXUrls(domain, timeout, recursive)
}

func GetUrlScanUrls(domain string, timeout int, apikey string) ([]string, error) {
  return core.GetUrlScanUrls(domain, timeout, apikey)
}


