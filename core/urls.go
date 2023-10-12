package core

import (
	"net/http"
	"sync"

	"github.com/D3Ext/go-recon/core/providers"
)

// main function to enumerate urls about provided domain, urls are sent through channel
// set "recursive" to false if you don't want to get urls related to subdomains
func GetAllUrls(domain string, results chan string, client *http.Client, recursive bool) error {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		providers.WaybackUrls(domain, results, client, 2, recursive)
		/*err := providers.WaybackUrls(domain, results, client, workers, recursive)
		  if err != nil {
		    return err
		  }*/
	}()

	go func() {
		defer wg.Done()
		providers.AlienVaultUrls(domain, results, client, recursive)
		/*err = providers.AlienVaultUrls(domain, results, client, recursive)
		  if err != nil {
		    return err
		  }*/
	}()

	wg.Wait()

	return nil
}
