package gorecon

import (
	"github.com/D3Ext/go-recon/core"
	p "github.com/D3Ext/go-recon/core/providers"
	"net/http"
)

// this function sents through provided channel all the gathered subdomains
// providers slice is used to configure the providers to use
// it also receives a client so you can custom most of the process
// Example: err := GetSubdomains("example.com", results, []string{"alienvault", "crt", "rapiddns", "wayback"}, gorecon.DefaultClient())
func GetSubdomains(domain string, results chan string, providers []string, client *http.Client) error {
	return core.GetSubdomains(domain, results, providers, client)
}

/*

providers functions

*/

func AlienVault(domain string, results chan string, client *http.Client) error {
	return p.AlienVault(domain, results, client)
}

func Anubis(domain string, results chan string, client *http.Client) error {
	return p.Anubis(domain, results, client)
}

func CommonCrawl(domain string, results chan string, client *http.Client) error {
	return p.CommonCrawl(domain, results, client)
}

func Crt(domain string, results chan string, client *http.Client) error {
	return p.Crt(domain, results, client)
}

func Digitorus(domain string, results chan string, client *http.Client) error {
	return p.Digitorus(domain, results, client)
}

func HackerTarget(domain string, results chan string, client *http.Client) error {
	return p.HackerTarget(domain, results, client)
}

func RapidDns(domain string, results chan string, client *http.Client) error {
	return p.RapidDns(domain, results, client)
}

func Wayback(domain string, results chan string, client *http.Client) error {
	return p.Wayback(domain, results, client)
}
