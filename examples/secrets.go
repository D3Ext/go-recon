package main

import (
  "fmt"
  "log"
  "github.com/D3Ext/go-recon/pkg/gorecon"
)

func main(){

  timeout := 4000 // in milliseconds

  secrets, err := gorecon.FindSecrets("https://example.com/endpoint.js", timeout) // return []string, error
  if err != nil {
    log.Fatal(err)
  }

  fmt.Println("Secrets:", secrets)
}
