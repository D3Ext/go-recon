package main

import (
  "fmt"
  "github.com/D3Ext/go-recon/pkg/gorecon"
)

func main(){

  results := make(chan string)

  timeout := 5000 // in milliseconds

  recursive := true // retrieve subdomains urls

  // func GetWaybackUrls(domain string, results chan string, timeout int, recursive bool) {}
  go gorecon.GetAllUrls("example.com", results, timeout, recursive)
  for url := range results {
    fmt.Println(url)
  }
}


