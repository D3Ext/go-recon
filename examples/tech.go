package main

import (
  "fmt"
  "log"
  "github.com/D3Ext/go-recon/pkg/gorecon"
)

func main(){
  timeout := 4000 // in milliseconds

  techs, err := gorecon.GetTech("https://github.com", timeout)
  if err != nil {
    log.Fatal(err)
  }

  fmt.Println(techs)
}


