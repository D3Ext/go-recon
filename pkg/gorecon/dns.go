package gorecon

import (
	"github.com/D3Ext/go-recon/core"
)

/*

// struct used by main function
type DnsInfo struct {
        Domain string   `json:"domain"` // given domain
        CNAME  string   `json:"cname"`  // returns the canonical name for the given host
        TXT    []string `json:"txt"`    // returns the DNS TXT records for the given domain name
        MX     []MX     `json:"mx"`     // returns a slice of MX (Mail eXchanges)
        NS     []NS     `json:"ns"`     // returns a slice of NS (Name Server)
        Hosts  []string `json:"hosts"`  // returns a slice of given host's IPv4 and IPv6 addresses
}

*/

// main function for DNS information gathering
// it receives a domain and tries to find most important info
// and returns a DnsInfo struct and an error
func Dns(domain string) (core.DnsInfo, error) {
	return core.Dns(domain)
}
