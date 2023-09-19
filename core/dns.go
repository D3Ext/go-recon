package core

import (
	"net"
)

// same as net.MX
type MX struct {
	Host string `json:"host"`
	Pref uint16 `json:"pref"`
}

// same as net.NS
type NS struct {
	Host string `json:"host"`
}

// struct used by main function
type DnsInfo struct {
	Domain string   `json:"domain"` // given domain
	CNAME  string   `json:"cname"`  // returns the canonical name for the given host
	TXT    []string `json:"txt"`    // returns the DNS TXT records for the given domain name
	MX     []MX     `json:"mx"`     //
	NS     []NS     `json:"ns"`     //
	Hosts  []string `json:"hosts"`  // returns a slice of given host's IPv4 and IPv6 addresses
}

func Dns(domain string) (DnsInfo, error) {
	var dns_info DnsInfo

	cname, err := net.LookupCNAME(domain)
	if err != nil {
		return dns_info, err
	}

	dns_info.CNAME = cname

	txt, err := net.LookupTXT(domain)
	if err != nil {
		return dns_info, err
	}

	dns_info.TXT = txt

	raw_ns, err := net.LookupNS(domain)
	if err != nil {
		return dns_info, err
	}

	var ns []NS

	for _, n := range raw_ns {
		ns = append(ns, NS{Host: n.Host})
	}

	dns_info.NS = ns

	raw_mx, err := net.LookupMX(domain)
	if err != nil {
		return dns_info, err
	}

	var mx []MX

	for _, m := range raw_mx {
		mx = append(mx, MX{Host: m.Host, Pref: m.Pref})
	}

	dns_info.MX = mx

	hosts, err := net.LookupHost(domain)
	if err != nil {
		return dns_info, err
	}

	dns_info.Hosts = hosts
	dns_info.Domain = domain

	return dns_info, nil
}
