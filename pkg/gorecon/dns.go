package gorecon

import (
	"github.com/D3Ext/go-recon/core"
)

/*

type DnsInfo struct {
  Domain      string    // given domain
  CNAME       string    // returns the canonical name for the given host
  TXT         []string  // returns the DNS TXT records for the given domain name
  MX          []MX //
  NS          []NS //
  Hosts       []string  // returns a slice of given host's IPv4 and IPv6 addresses
}

*/

func Dns(domain string) (core.DnsInfo, error) {
	return core.Dns(domain)
}
