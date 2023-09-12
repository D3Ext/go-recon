package main

import (
  "fmt"
  "github.com/D3Ext/go-recon/pkg/gorecon"
)

func main(){
  urls := []string{"http://example.com/index.php", "http://example.com/index.php", "http://example.com/index.php?id=1", "http://example.com/index.php?id=2", "http://example.com/index.php?id=1&page=3"}

  filters := []string{""} // "vuln", "hasparams", "noparams", "hasextension", "noextension", "nocontent"

  new_urls := gorecon.FilterUrls(urls, filters)
  fmt.Println(new_urls)
}

