package gorecon

import (
	"github.com/D3Ext/go-recon/core"
)

// results := make(chan string)
// go gorecon.GetEndpointsFromFile(urls, results, 15, 5000)
//  for endpoint := range results {
//    fmt.Println(endpoint)
//  }

func GetEndpoints(urls []string, results chan string, workers int, timeout int) {
	core.GetEndpoints(urls, results, workers, timeout)
}

func FetchEndpoints(urls <-chan string, results chan string, user_agent string, proxy string, timeout int) {
	core.FetchEndpoints(urls, results, user_agent, proxy, timeout)
}
