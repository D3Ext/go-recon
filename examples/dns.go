package main

import (
  "fmt"
  "log"
  "github.com/D3Ext/go-recon/pkg/gorecon"
)

func main(){
  dns_info, err := gorecon.Dns("hackthebox.com")
  if err != nil {
    log.Fatal(err)
  }

/*

type DnsInfo struct {
  Domain      string    // given domain
  CNAME       string    // returns the canonical name for the given host
  TXT         []string  // returns the DNS TXT records for the given domain name
  MX          []*net.MX //
  NS          []*net.NS //
  Hosts       []string  // returns a slice of given host's IPv4 and IPv6 addresses
}

*/

  fmt.Println(dns_info)
}

