package core

import (
  "net"
  "time"
  "net/url"
  "net/http"
  "crypto/tls"
)

func CreateHttpClient(timeout int) (*http.Client) {
  // Create requests client
  t := time.Duration(timeout) * time.Millisecond

  var transport = &http.Transport{
    MaxIdleConns:      30,
    IdleConnTimeout:   time.Second,
    DisableKeepAlives: true,
    TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, // Disable ssl verify
    DialContext: (&net.Dialer{
      Timeout:   t,
      KeepAlive: time.Second,
    }).DialContext,
  }

  redirect := func(req *http.Request, via []*http.Request) error {
    return http.ErrUseLastResponse // Don't follow redirect
  }

  client := &http.Client{ // Create requests client
    Transport:     transport,
    CheckRedirect: redirect,
    Timeout:       t,
  }

  return client
}

func CreateHttpClientWithProxy(timeout int, proxy string) (*http.Client, error) {
  // Create requests client
  t := time.Duration(timeout) * time.Millisecond

  proxy_url, err := url.Parse(proxy)
  if err != nil {
    return &http.Client{}, err
  }

  var transport = &http.Transport{
    Proxy:             http.ProxyURL(proxy_url),
    MaxIdleConns:      30,
    IdleConnTimeout:   time.Second,
    DisableKeepAlives: true,
    TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, // Disable ssl verify
    DialContext: (&net.Dialer{
      Timeout:   t,
      KeepAlive: time.Second,
    }).DialContext,
  }

  redirect := func(req *http.Request, via []*http.Request) error {
    return http.ErrUseLastResponse // Don't follow redirect
  }

  client := &http.Client{ // Create requests client
    Transport:     transport,
    CheckRedirect: redirect,
    Timeout:       t,
  }

  return client, nil
}



