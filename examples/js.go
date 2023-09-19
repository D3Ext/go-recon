package main

import (
	"fmt"
	"github.com/D3Ext/go-recon/pkg/gorecon"
)

func main() {

	urls := []string{"https://example.com", "https://github.com", "https://hackthebox.com", "https://hackerone.com"}

	results := make(chan string)

	workers := 15 // create 15 concurrent workers

	timeout := 5000 // in milliseconds

	//func GetEndpoints(urls []string, results chan string, workers int, timeout int) {}
	go gorecon.GetEndpoints(urls, results, workers, timeout)
	for endpoint := range results {
		fmt.Println(endpoint)
	}
}
