package main

import (
  "fmt"
  "log"
  "github.com/D3Ext/go-recon/pkg/gorecon"
)

func main(){

  url := "http://example.com/dir/"

  word := "secret"

  timeout := 5000

  urls, status_codes, err := gorecon.Check403(url, word, timeout)
  if err != nil {
    log.Fatal(err)
  }

  for i, _ := range urls {
    fmt.Println("Url:", urls[i])
    fmt.Println("Status code:", status_codes[i])
  }
}

