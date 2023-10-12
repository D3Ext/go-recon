package core

import (
	"github.com/likexian/whois"
	wp "github.com/likexian/whois-parser"
)

// send WHOIS query to given domain to retrieve public info
// Example: info, err := gorecon.Whois("hackthebox.com")
func Whois(domain string) (wp.WhoisInfo, error) {
	raw, err := whois.Whois(domain)
	if err != nil {
		return wp.WhoisInfo{}, err
	}

	result, err := wp.Parse(raw)
	if err != nil {
		return result, err
	}

	return result, nil
}
