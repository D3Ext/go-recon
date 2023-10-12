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

	client := &http.Client{ // Create requests client
		Timeout: 10000 * time.Millisecond,
	}

	// use all providers
	providers := []string{"alienvault", "anubis", "commoncrawl", "crt", "digitorus", "hackertarget", "rapiddns", "wayback"}

	go func() {
		for res := range results {
			fmt.Println(res)
		}
	}()

	// Get all subdomains from given providers (subdomains channel is closed when all subdomains are enumerated)
	fmt.Println("[*] Retrieving subdomains from all providers:")
	err := gorecon.GetSubdomains("hackthebox.com", results, providers, client)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n[*] Retrieving subdomains from Crt.sh:")
	err = gorecon.Crt("hackthebox.com", results, client)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n[*] Retrieving subdomains from HackerTarget:")
	err = gorecon.HackerTarget("hackthebox.com", results, client)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n[*] Retrieving subdomains from AlienVault:")
	err = gorecon.AlienVault("hackthebox.com", results, client)
	if err != nil {
		log.Fatal(err)
	}

	// do the same for the other providers
}
