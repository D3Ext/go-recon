package core

import (
  "net"
  "time"
  "regexp"
  "strconv"
  "strings"
  "net/http"
  "io/ioutil"
  "crypto/tls"
  "encoding/json"
)

func DetectWaf(url string, payload string, keyword string, timeout int) (string, error) {
  if (payload == "") {
    payload = "../../../../../etc/passwd"
  }

  if (timeout == 0) {
    timeout = 4000
  }

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

  client := &http.Client{ // Create requests client
    Transport:     transport,
    Timeout:       t,
  }

  var json_url string = "https://raw.githubusercontent.com/D3Ext/AORT/main/utils/wafsign.json"
  var m map[string]interface{}

  req, _ := http.NewRequest("GET", json_url, nil)
  req.Header.Add("Connection", "keep-alive")
  req.Close = true

  resp, err := client.Do(req) // Send request
  if err != nil {
    return "", err
  }
  defer resp.Body.Close()

  // Read raw response
  data, _ := ioutil.ReadAll(resp.Body)

  // Parse json
  err = json.Unmarshal(data, &m)
  if err != nil {
    return "", nil
  }

  var payload_url string
  if (keyword != "") {
    payload_url = strings.Replace(url, keyword, payload, -1)
  } else {
    if strings.HasSuffix(payload_url, "/") {
      payload_url = url + payload
    } else {
      payload_url = url + "/" + payload
    }
  }

  req, _ = http.NewRequest("GET", payload_url, nil)
  req.Header.Add("Connection", "close")
  req.Close = true

  resp, err = client.Do(req) // Send request
  if err != nil {
    return "", err
  }
  defer resp.Body.Close()

  // Define some values which are compared with WAF vendors data to detect them
  var cookies []string
  var headers []string

  code := strconv.Itoa(resp.StatusCode) // Get status code
  page, _ := ioutil.ReadAll(resp.Body) // Parse page content
  for _, cookie := range resp.Cookies() { // Parse cookie names and values to compare them later
    cookies = append(cookies, cookie.Name)
    cookies = append(cookies, cookie.Value)
  }

  for key, values := range resp.Header { // Parse headers
    headers = append(headers, key)
    for _, v := range values {
      headers = append(headers, v)
    }
  }

  var result float32 = 0
  for key, value := range m { // iterate over json data
    code_to_check, _ := find(value, "code")
    if (code_to_check.(string) != "") {
      res, err := regexp.MatchString(code_to_check.(string), code)
      if err != nil {
        continue
      }

      if (res) {
        result += 0.5
      }
    }

    page_to_check, _ := find(value, "page")
    if (page_to_check.(string) != "") {
      res, err := regexp.MatchString(page_to_check.(string), string(page))
      if err != nil {
        continue
      }

      if (res) {
        result += 1
      }
    }

    headers_to_check, _ := find(value, "headers")
    if (headers_to_check.(string) != "") {
      for _, h := range headers {
        res, err := regexp.MatchString(headers_to_check.(string), h)
        if err != nil {
          continue
        }

        if (res) {
          result += 1
        }
      }
    }

    cookies_to_check, _ := find(value, "cookie")
    if (cookies_to_check.(string) != "") {
      for _, c := range cookies {
        res, err := regexp.MatchString(cookies_to_check.(string), c)
        if err != nil {
          continue
        }

        if (res) {
          result += 1
        }
      }
    }

    if (result >= 1) {
      return key, nil
    }

    result = 0
  }

  return "", nil
}

func find(o interface{}, key string) (interface{}, bool) { // Function used to find key values
  //if the argument is not a map, ignore it
  mobj, ok := o.(map[string]interface{})
  if !ok {
    return nil, false
  }

  for k, v := range mobj {
    // key match, return value
    if k == key {
      return v, true
    }

    // if the value is a map, search recursively
    if m, ok := v.(map[string]interface{}); ok {
      if res, ok := find(m, key); ok {
        return res, true
      }
    }

    // if the value is an array, search recursively 
    // from each element
    if va, ok := v.([]interface{}); ok {
      for _, a := range va {
        if res, ok := find(a, key); ok {
          return res, true
        }
      }
    }
  }

  // element not found
  return nil, false
}



