package main

import (
	"fmt"
	"github.com/D3Ext/go-recon/pkg/gorecon"
	"log"
)

func main() {

	url := "http://example.com/dir/"

	word := "secret"

	timeout := 5000

	urls, status_codes, err := gorecon.Check403(url, word, timeout)
	if err != nil {
		log.Fatal(err)
	}

	for i := range urls {
		fmt.Println("Url:", urls[i])
		fmt.Println("Status code:", status_codes[i])
	}
}
