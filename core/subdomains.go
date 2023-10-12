package core

import (
	p "github.com/D3Ext/go-recon/core/providers"
	"net/http"
	"sync"
)

// this function sents through provided channel all the gathered subdomains
// providers slice is used to configure the providers to use
// it also receives a client so you can custom most of the process
// Example: err := GetSubdomains("example.com", results, []string{"alienvault", "crt", "rapiddns", "wayback"}, gorecon.DefaultClient())
func GetSubdomains(dom string, results chan string, providers []string, client *http.Client) error {
	var wg sync.WaitGroup
	wg.Add(len(providers))

	if stringInSlice("alienvault", providers) {
		go func() {
			defer wg.Done()

			p.AlienVault(dom, results, client)
			/*err := p.AlienVault(dom, results, client)
			  if err != nil {
			    return err
			  }*/
		}()
	}

	if stringInSlice("anubis", providers) {
		go func() {
			defer wg.Done()

			p.Anubis(dom, results, client)
			/*err := p.Anubis(dom, results, client)
			  if err != nil {
			    return err
			  }*/
		}()
	}

	if stringInSlice("commoncrawl", providers) {
		go func() {
			defer wg.Done()

			p.CommonCrawl(dom, results, client)
			/*err := p.CommonCrawl(dom, results, client)
			  if err != nil {
			    return err
			  }*/
		}()
	}

	if stringInSlice("crt", providers) || stringInSlice("crtsh", providers) {
		go func() {
			defer wg.Done()

			p.Crt(dom, results, client)
			/*err := p.Crt(dom, results, client)
			  if err != nil {
			    return err
			  }*/
		}()
	}

	if stringInSlice("digitorus", providers) {
		go func() {
			defer wg.Done()

			p.Digitorus(dom, results, client)
			/*err := p.Digitorus(dom, results, client)
			  if err != nil {
			    return err
			  }*/
		}()
	}

	if stringInSlice("hackertarget", providers) {
		go func() {
			defer wg.Done()

			p.HackerTarget(dom, results, client)
			/*err := p.HackerTarget(dom, results, client)
			  if err != nil {
			    return err
			  }*/
		}()
	}

	if stringInSlice("rapiddns", providers) {
		go func() {
			defer wg.Done()

			p.RapidDns(dom, results, client)
			/*err := p.RapidDns(dom, results, client)
			  if err != nil {
			    return err
			  }*/
		}()
	}

	if stringInSlice("wayback", providers) {
		go func() {
			defer wg.Done()

			p.Wayback(dom, results, client)
			/*err := p.Wayback(dom, results, client)
			  if err != nil {
			    return err
			  }*/
		}()
	}

	wg.Wait()

	return nil
}
