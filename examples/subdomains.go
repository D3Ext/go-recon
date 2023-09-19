package main

import (
	"fmt"
	"github.com/D3Ext/go-recon/pkg/gorecon"
	"log"
)

func main() {
	subdomains := make(chan string)
	timeout := 10000 // in milliseconds

	// Get all subdomains using Crt, HackerTarget and AlienVault
	go gorecon.GetAllSubdomains("hackthebox.com", subdomains, timeout)
	for sub := range subdomains {
		fmt.Println(sub)
	}

	fmt.Println("-----------")

	subs1, err := gorecon.Crt("hackthebox.com", timeout)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(subs1)

	fmt.Println("-----------")

	subs2, err := gorecon.HackerTarget("hackthebox.com", timeout)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(subs2)

	fmt.Println("-----------")

	subs3, err := gorecon.AlienVault("hackthebox.com", timeout)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(subs3)
}
