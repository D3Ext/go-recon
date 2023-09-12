package gorecon

import (
  "net/http"
  "github.com/D3Ext/go-recon/core"
)

func CheckRedirect(url string, c *http.Client, payloads []string, keyword string) ([]string, error) {
  return core.CheckRedirect(url, c, payloads, keyword)
}

func GetPayloads() []string {
  return core.GetPayloads()
}

