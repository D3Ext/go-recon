package main

import (
  "fmt"
  "log"
  "time"
  "net/http"
  "github.com/D3Ext/go-recon/pkg/gorecon"
)

func main(){

  client := &http.Client{
    Timeout: time.Duration(5000 * time.Millisecond),
  }

  urls, err := gorecon.CheckRedirect("https://example.com/?url=FUZZ", client, gorecon.GetPayloads(), "FUZZ")
  if err != nil {
    log.Fatal(err)
  }

  fmt.Println(urls)
}
