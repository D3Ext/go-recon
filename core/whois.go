package core

import (
	"github.com/likexian/whois"
	wp "github.com/likexian/whois-parser"
)

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
