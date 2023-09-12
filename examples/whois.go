package main

import (
  "fmt"
  "log"
  "github.com/D3Ext/go-recon/pkg/gorecon"
)

func main(){

  results, err := gorecon.Whois("example.com")
  if err != nil {
    log.Fatal(err)
  }

  fmt.Println(results.Domain.Name)
  fmt.Println(results.Domain.NameServers)
  fmt.Println(results.Domain.ID)
}
