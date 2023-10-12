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
		Timeout: 5000 * time.Millisecond,
	}

	// try to detect waf with default payload (../../../../../etc/passwd)
	waf, err := gorecon.DetectWaf("https://hackthebox.com", "", "", client)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("WAF:", waf)

	// replace given keyword (FUZZ) with given payload (' or 1=1-- -) on given url
	waf, err = gorecon.DetectWaf("https://hackthebox.com/index.php?foo=FUZZ", "FUZZ", "' or 1=1-- -", client)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("WAF:", waf)
}
