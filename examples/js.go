package main

import (
	"fmt"
	"github.com/D3Ext/go-recon/pkg/gorecon"
	"log"
	"net/http"
	"time"
)

func main() {

	urls := []string{"https://example.com", "https://github.com", "https://hackthebox.com", "https://hackerone.com"}

	results := make(chan string)

	workers := 4 // create 15 concurrent workers

	client := &http.Client{
		Timeout: 5000 * time.Millisecond,
	}

	go func() {
		for endpoint := range results {
			fmt.Println(endpoint)
		}
	}()

	//func GetEndpoints(urls []string, results chan string, workers int, client *http.Client) {}
	err := gorecon.GetEndpoints(urls, results, workers, client)
	if err != nil {
		log.Fatal(err)
	}
}
