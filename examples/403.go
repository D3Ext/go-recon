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

  client := gorecon.CreateHttpClient(timeout)

	urls, status_codes, err := gorecon.Check403(url, word, client, "Mozilla/5.0 (X11; Linux x86_64; rv:78.0) Gecko/20100101 Firefox/78.0")
	if err != nil {
		log.Fatal(err)
	}

	for i := range urls {
		fmt.Println("Url:", urls[i])
		fmt.Println("Status code:", status_codes[i])
	}
}
