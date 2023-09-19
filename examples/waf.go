package main

import (
	"fmt"
	"github.com/D3Ext/go-recon/pkg/gorecon"
	"log"
)

func main() {
	timeout := 5000 // in milliseconds

	waf, err := gorecon.DetectWaf("https://hackthebox.com", "", "", timeout)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("WAF:", waf)
}
