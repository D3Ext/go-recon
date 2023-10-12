package main

import (
	"fmt"
	"github.com/D3Ext/go-recon/pkg/gorecon"
	"log"
	"net/http"
	"time"
)

func main() {

	results := make(chan string)

	client := &http.Client{
		Timeout: 10000 * time.Millisecond,
	}

	recursive := true // retrieve subdomains urls

	go func() {
		for url := range results {
			fmt.Println(url)
		}
	}()

	err := gorecon.GetAllUrls("example.com", results, client, recursive)
	if err != nil {
		log.Fatal(err)
	}
}
