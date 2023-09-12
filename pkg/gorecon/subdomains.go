package gorecon

import (
  "github.com/D3Ext/go-recon/core"
)

// results := make(chan string)
// go gorecon.GetAllSubdomains("example.com", results, 10000)
func GetAllSubdomains(domain string, subdomains chan string, timeout int) {
  core.GetAllSubdomains(domain, subdomains, timeout)
}

// subdomains, err := gorecon.Crt("example.com", 8000)
func Crt(domain string, timeout int) ([]string, error) {
  return core.Crt(domain, timeout)
}

// subdomains, err := gorecon.HackerTarget("example.com", 8000)
func HackerTarget(domain string, timeout int) ([]string, error) {
  return core.HackerTarget(domain, timeout)
}

// subdomains, err := gorecon.AlienVault("example.com", 8000)
func AlienVault(domain string, timeout int) ([]string, error) {
  return core.AlienVault(domain, timeout)
}

