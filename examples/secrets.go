package main

import (
	"fmt"
	"github.com/D3Ext/go-recon/pkg/gorecon"
	"log"
	"net/http"
	"time"
)

func main() {

	client := &http.Client{
		Timeout: 4000 * time.Millisecond,
	}

	secrets, err := gorecon.FindSecrets("https://example.com/endpoint.js", client) // return []string, error
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Secrets:", secrets)
}
