package core

import (
  "time"
  "sync"
  "errors"
  "strconv"
  "strings"
  "net/url"
  "net/http"
  "io/ioutil"
  "crypto/tls"
  "encoding/json"
)

type otxResult struct {
  HasNext    bool `json:"has_next"`
  ActualSize int  `json:"actual_size"`
  URLList    []struct {
    Domain   string `json:"domain"`
    URL      string `json:"url"`
    Hostname string `json:"hostname"`
    HTTPCode int    `json:"httpcode"`
    PageNum  int    `json:"page_num"`
    FullSize int    `json:"full_size"`
    Paged    bool   `json:"paged"`
  } `json:"url_list"`
}

type SearchResults struct {
  Results []struct {
    Task struct {
      Visibility string    `json:"visibility"`
      Method     string    `json:"method"`
      Time       time.Time `json:"time"`
      Source     string    `json:"source"`
      URL        string    `json:"url"`
    } `json:"task"`
    Stats struct {
      UniqIPs           int `json:"uniqIPs"`
      ConsoleMsgs       int `json:"consoleMsgs"`
      DataLength        int `json:"dataLength"`
      EncodedDataLength int `json:"encodedDataLength"`
      Requests          int `json:"requests"`
    } `json:"stats"`
    Page struct {
      Country string `json:"country"`
      Server  string `json:"server"`
      City    string `json:"city"`
      Domain  string `json:"domain"`
      IP      string `json:"ip"`
      Asnname string `json:"asnname"`
      Asn     string `json:"asn"`
      URL     string `json:"url"`
      Ptr     string `json:"ptr"`
    } `json:"page"`
    UniqCountries int    `json:"uniq_countries"`
    ID            string `json:"_id"`
    Result        string `json:"result"`
  } `json:"results"`
  Total int `json:"total"`
}

var wg sync.WaitGroup

func GetAllUrls(domain string, results chan string, timeout int, recursive bool) {

  wg.Add(1)
  go GetWaybackUrls(domain, results, timeout, recursive) // launch as goroutine because it's a long process and sends tons of requests

  subs2, _ := GetOTXUrls(domain, timeout, recursive)
  for _, s := range subs2 {
    results <- s
  }

  wg.Wait()
  close(results)
}

// too much content so apply high timeout
func GetWaybackUrls(domain string, results chan string, timeout int, recursive bool) error {

  // Create requests client
  client := CreateHttpClient(timeout)

  // First we get the max number of pages
  var max_page int
  var err error
  var page_url string

  if (!recursive) {
    page_url = "http://web.archive.org/cdx/search/cdx?url=" + domain + "/*&showNumPages=true"
  } else {
    page_url = "http://web.archive.org/cdx/search/cdx?url=*." + domain + "/*&showNumPages=true"
  }

  req, err := http.NewRequest("GET", page_url, nil)
  if err != nil {
    return errors.New("WaybackMachine seems to be slow or down!")
  }

  res, err := client.Do(req)
  if err != nil {
    return errors.New("WaybackMachine seems to be slow or down!")
  }
  defer res.Body.Close()

  raw, err := ioutil.ReadAll(res.Body)
  if err != nil {
    return errors.New("WaybackMachine seems to be slow or down!")
  }

  max_page, err = strconv.Atoi(strings.TrimSuffix(string(raw), "\n"))
  if err != nil {
    return errors.New("WaybackMachine seems to be slow or down!")
  }

  var current_url string
  // Now iterate over numbers from 0 to max page number
  for i := 0; i < max_page; i++ {

    if (!recursive) {
      current_url = "http://web.archive.org/cdx/search/cdx?url=" + domain + "/*&output=json&collapse=urlkey&page=" + strconv.Itoa(i)
    } else {
      current_url = "http://web.archive.org/cdx/search/cdx?url=*." + domain + "/*&output=json&collapse=urlkey&page=" + strconv.Itoa(i)
    }

    req, err := http.NewRequest("GET", current_url, nil)
    if err != nil {
      continue
    }

    res, err := client.Do(req)
    if err != nil {
      continue
    }
    defer res.Body.Close()

    raw, err := ioutil.ReadAll(res.Body)
    if err != nil {
      continue
    }

    var wrapper [][]string
    err = json.Unmarshal(raw, &wrapper)
    if err != nil {
      continue
    }

    skip := true
    for _, urls := range wrapper {
      if skip {
        skip = false
        continue
      }

      results <- urls[2]
    }
  }

  wg.Done()
  return nil
}

func GetOTXUrls(domain string, timeout int, recursive bool) ([]string, error) {
  base_url := "https://otx.alienvault.com/api/v1/indicators/domain/" + domain + "/url_list?limit=100&page="
  out := []string{}

  for i := 1; i < 100; i++ {
    req, err := http.NewRequest("GET", base_url + strconv.Itoa(i), nil)
    if err != nil {
      continue
    }

    c := &http.Client{
      Timeout:   time.Duration(timeout * 1000000),
      Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
    }

    resp, err := c.Do(req)
    if err != nil {
      continue
    }
    defer resp.Body.Close()

    raw, err := ioutil.ReadAll(resp.Body)
    if err != nil {
      return nil, err
    }

    var result otxResult
    err = json.Unmarshal(raw, &result)
    if err != nil {
      return nil, err
    }

    if (result.HasNext == false) {
      break
    }

    for _, e := range result.URLList {

      if (recursive == false) {
        current_url, err := url.Parse(e.URL)
        if err != nil {
          continue
        }
        hostname := strings.TrimPrefix(current_url.Hostname(), "www.")

        if (hostname == domain) {
          out = append(out, e.URL)
        }

      } else {
        out = append(out, e.URL)
      }
    }
  }

  return out, nil
}

func GetUrlScanUrls(domain string, timeout int, apikey string) ([]string, error) {
  base_url := "https://urlscan.io/api/v1/search/?q=domain:" + domain + "&size=10000"
  out := []string{}

  req, err := http.NewRequest("GET", base_url, nil)
  if err != nil {
    return []string{""}, err
  }

  req.Header.Set("API-Key", apikey)

  c := &http.Client{
    Timeout:   time.Duration(timeout * 1000000),
    Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
  }

  resp, err := c.Do(req)
  if err != nil {
    return nil, err
  }
  defer resp.Body.Close()

  var results SearchResults
  
  raw, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return nil, err
  }

  err = json.Unmarshal(raw, &results)
  if err != nil {
    return nil, err
  }

  for _, e := range results.Results {
    out = append(out, e.Task.URL)
  }

  return out, nil
}

