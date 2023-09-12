package main

import (
  "os"
  "fmt"
  "log"
  "flag"
  "sync"
  "bufio"
  "regexp"
  "strings"
  "strconv"
  "net/http"
  "io/ioutil"
  nu "net/url"
  "encoding/json"
  "github.com/fatih/color"

  "github.com/D3Ext/go-recon/core"
)

var red func(a ...interface{}) string = color.New(color.FgRed).SprintFunc()
var cyan func(a ...interface{}) string = color.New(color.FgCyan).SprintFunc()
var blue func(a ...interface{}) string = color.New(color.FgBlue).SprintFunc()
var green func(a ...interface{}) string = color.New(color.FgGreen).SprintFunc()
var magenta func(a ...interface{}) string = color.New(color.FgMagenta).SprintFunc()
var yellow func(a ...interface{}) string = color.New(color.FgYellow).SprintFunc()

func helpPanel() {
  fmt.Println(`Usage of gr-waf:
    -u)       url to identify its WAF (i.e. https://example.com)
    -l)       file containing a list of urls to identify their WAFs (one url per line)
    -p)       payload used to trigger WAF (default=../../../../../etc/passwd)
    -k)       keyword to replace in urls (if url doesn't contain keyword nothing is changed) (default=FUZZ)
    -w)       number of concurrent workers (default=15)
    -a)       user agent to include on requests
    -c)       print colors on output (recommended)
    -t)       milliseconds to wait before timeout (default=4000)
    -q)       don't print banner nor logging, only output
    -h)       print help panel
  
Examples:
    gr-waf -u https://example.com -c
    gr-waf -u https://example.com/index.php?foo=FUZZ -k FUZZ -q
    gr-waf -u https://example.com/index.php?foo=FUZZ -p "' or 1=1"
    cat urls.txt | gr-waf -k TEST
    `)
}

func main(){
  var url string
  var list string
  var payload string
  var keyword string
  var workers int
  var timeout int
  var user_agent string
  var use_color bool
  var quiet bool
  var help bool
  var stdin bool

  flag.StringVar(&url, "u", "", "url to identify its WAF (i.e. https://example.com)")
  flag.StringVar(&list, "l", "", "file containing a list of urls to identify their WAFs (one url per line)")
  flag.StringVar(&payload, "p", "../../../../../etc/passwd", "payload used to trigger WAF")
  flag.StringVar(&keyword, "k", "FUZZ", "keyword to replace in urls (if url doesn't contain keyword nothing is changed) (default=FUZZ)")
  flag.IntVar(&workers, "w", 15, "number of concurrent workers")
  flag.IntVar(&timeout, "t", 4000, "milliseconds to wait before timeout")
  flag.StringVar(&user_agent, "a", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/93.0.4577.63 Safari/537.36", "user agent to include on requests")
  flag.BoolVar(&use_color, "c", false, "print colors on output")
  flag.BoolVar(&quiet, "q", false, "don't print banner nor logging, only output")
  flag.BoolVar(&help, "h", false, "print help panel")
  flag.Parse()

  t1 := core.StartTimer()

  if (!quiet) {
    fmt.Println(core.Banner())
  }

  if (help) {
    helpPanel()
    os.Exit(0)
  }

  // Check if stdin has value
  fi, err := os.Stdin.Stat()
  if err != nil {
    log.Fatal(err)
  }

  if fi.Mode() & os.ModeNamedPipe == 0 {
    stdin = false // stdin is empty
  } else {
    stdin = true // stdin has value
  }

  // if domain, list and stdin parameters are empty print help panel and exit
  if (url == "") && (list == "") && (!stdin) {
    fmt.Println()
    helpPanel()
    os.Exit(0)
  }

  if (url != "") && (list != "") {
    helpPanel()
    core.Red("You can't use (-u) and (-l) at same time", use_color)
    os.Exit(0)
  }

  client := core.CreateHttpClient(timeout)

  var json_url string = "https://raw.githubusercontent.com/D3Ext/AORT/main/utils/wafsign.json"
  //var json_url string = "https://raw.githubusercontent.com/D3Ext/go-recon/main/utils/waf_vendors.json"
  var m map[string]interface{}

  if (url != "") {

    // send request to waf vendors data (json format)
    req, _ := http.NewRequest("GET", json_url, nil)
    req.Header.Add("Connection", "keep-alive")
    req.Header.Set("User-Agent", user_agent)
    req.Close = true

    resp, err := client.Do(req) // Send request
    if err != nil {
      log.Fatal(err)
    }
    defer resp.Body.Close()

    // Read raw response
    data, _ := ioutil.ReadAll(resp.Body)

    if (!quiet) {
      if (use_color) {
        fmt.Println("[" + magenta("*") + "] Parsing WAF vendors data... (" + cyan(strconv.Itoa(len(strings.Split(string(data), "\n")))) + " lines)")
      } else {
        fmt.Println("[*] Parsing WAF vendors data... (" + strconv.Itoa(len(strings.Split(string(data), "\n"))) + " lines)")
      }
    }

    // Parse json
    err = json.Unmarshal(data, &m)
    if err != nil {
      log.Fatal(err)
    }

    // Parse given url
    url_parse, err := nu.Parse(url)
    if err != nil {
      log.Fatal(err)
    }

    // Parse parameters
    params, err := nu.ParseQuery(url_parse.RawQuery)
    if err != nil {
      log.Fatal(err)
    }
    
    var param_to_test string
    for key, value := range params { // Check if any of parameter values equals keyword to replace with payload
      if (value[0] == keyword) {
        param_to_test = key
        break
      }
    }

    if (!quiet) {
      if (param_to_test == "") { // enter here if no parameter to test is given
        if (use_color) {
          fmt.Println("[" + red("!") + "] No parameters detected so analysis may not be accurated")
        } else {
          fmt.Println("[!] No parameters detected so analysis may not be accurated")
        }
      } else {
        if (use_color) {
          fmt.Println("[" + green("+") + "] Parameter to test detected:", cyan(param_to_test))
        } else {
          fmt.Println("[+] Parameter to test detected:", param_to_test)
        }
      }

    // Check if url is up
      core.Magenta("Testing connection with target url...", use_color)
    }

    req, _ = http.NewRequest("GET", url, nil)
    req.Header.Add("Connection", "keep-alive")
    req.Header.Set("User-Agent", user_agent)
    req.Close = true

    resp, err = client.Do(req) // Send request
    if err != nil {
      log.Fatal(err)
    }

    if (!quiet) {
      core.Green("Connection succeeded", use_color)
      core.Magenta("User-Agent: " + user_agent, use_color)
    }

    var payload_url string // Define url with payload
    if (strings.HasSuffix(url, "/")) && (param_to_test == "") {
      payload_url = url + payload

    } else if (!strings.HasSuffix(url, "/")) && (param_to_test == "") {
      payload_url = url + "/" + payload

    } else if (!strings.HasSuffix(url, "/")) && (param_to_test != "") {
      payload_url = strings.ReplaceAll(url, keyword, payload)
    }

    if (!quiet) {
      if (use_color) {
        fmt.Println("[" + magenta("*") + "] Payload url:", cyan(payload_url))
      } else {
        fmt.Println("[*] Payload url:", payload_url)
      }
    }

    req, _ = http.NewRequest("GET", payload_url, nil)
    req.Header.Add("Connection", "close")
    req.Header.Set("User-Agent", user_agent)
    req.Close = true

    resp, err = client.Do(req) // Send request
    if err != nil {
      log.Fatal(err)
    }
    defer resp.Body.Close()

    if (!quiet) {
      if (use_color) {
        if (resp.StatusCode >= 400) {
          fmt.Println("[" + red("!") + "] Status code:", cyan(resp.StatusCode))
        } else if (resp.StatusCode >= 300 && resp.StatusCode < 400) {
          fmt.Println("[" + blue("*") + "] Status code:", cyan(resp.StatusCode))
        } else if (resp.StatusCode >= 200 && resp.StatusCode < 300) {
          fmt.Println("[" + green("+") + "] Status code:", cyan(resp.StatusCode))
        }
        fmt.Println("[" + magenta("*") + "] Comparing values... (headers, response, cookies, status code)")
      } else {
        if (resp.StatusCode >= 400) {
          fmt.Println("[!] Status code:", resp.StatusCode)
        } else if (resp.StatusCode >= 300 && resp.StatusCode < 400) {
          fmt.Println("[*] Status code:", resp.StatusCode)
        } else if (resp.StatusCode >= 200 && resp.StatusCode < 300) {
          fmt.Println("[+] Status code:", resp.StatusCode)
        }
        fmt.Println("[*] Comparing values... (headers, response, cookies, status code)")
      }
    }

    // Define some values which are compared with WAF vendors data to detect them
    var cookies []string
    var headers []string

    var nl_prefix string
    if (!quiet) { // if quiet, don't print new line to maintain format
      nl_prefix = "\n"
    }

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
      code_to_check, _ := Find(value, "code")
      if (code_to_check.(string) != "") {
        res, err := regexp.MatchString(code_to_check.(string), code)
        if err != nil {
          continue
        }

        if (res) {
          result += 0.5
        }
      }

      page_to_check, _ := Find(value, "page")
      if (page_to_check.(string) != "") {
        res, err := regexp.MatchString(page_to_check.(string), string(page))
        if err != nil {
          continue
        }

        if (res) {
          result += 1
        }
      }

      headers_to_check, _ := Find(value, "headers")
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

      cookies_to_check, _ := Find(value, "cookie")
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
        if (use_color) {
          fmt.Println(nl_prefix + "[" + green("+") + "] WAF found:", cyan(key))
        } else {
          fmt.Println(nl_prefix + "[+] WAF found:", key)
        }

        break
      }

      result = 0
    }

    if (result == 0) {
      if (use_color) {
        fmt.Println(nl_prefix + "[" + red("-") + "] WAF not found")
      } else {
        fmt.Println(nl_prefix + "[-] WAF not found")
      }
    }

  } else if (list != "") || (stdin) {

    var f *os.File
    var err error
    if (list != "") { // get file descriptor from .txt file or stdin
      f, err = os.Open(list)
      if err != nil {
        log.Fatal(err)
      }
      defer f.Close()

    } else if (stdin) {
      f = os.Stdin
    }

    urls_c := make(chan string)
    var wg sync.WaitGroup

    for i := 0; i < workers; i++ { // create n workers
      wg.Add(1)

      go func(){
        for u := range urls_c { // get url from channel

          if (!strings.HasPrefix(u, "http://")) && (!strings.HasPrefix(u, "https://")) {
            url = "https://" + u
          }

          if (strings.HasSuffix(u, "=")) {
            url = u + payload

          } else if (!strings.HasSuffix(u, "/")) {
            url = u + "/" + payload
          }

          req, _ := http.NewRequest("GET", url, nil)
          req.Header.Add("Connection", "close")
          req.Header.Set("User-Agent", user_agent)
          req.Close = true

          resp, err := client.Do(req) // Send request
          if err != nil {
            log.Fatal(err)
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
            code_to_check, _ := Find(value, "code")
            if (code_to_check.(string) != "") {
              res, err := regexp.MatchString(code_to_check.(string), code)
              if err != nil {
                continue
              }

              if (res) {
                result += 0.5
              }
            }

            page_to_check, _ := Find(value, "page")
            if (page_to_check.(string) != "") {
              res, err := regexp.MatchString(page_to_check.(string), string(page))
              if err != nil {
                continue
              }

              if (res) {
                result += 1
              }
            }

            headers_to_check, _ := Find(value, "headers")
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

            cookies_to_check, _ := Find(value, "cookie")
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
              fmt.Println(u, "-", key)
              break
            }

            result = 0
          }

          if (result == 0) {
            fmt.Println(u, "-", "Not found")
          }
        }

        wg.Done() // finish worker
      }()
    }

    // send request to waf vendors data (json format)
    req, _ := http.NewRequest("GET", json_url, nil)
    req.Header.Add("Connection", "keep-alive")
    req.Header.Set("User-Agent", user_agent)
    req.Close = true

    resp, err := client.Do(req) // Send request
    if err != nil {
      log.Fatal(err)
    }
    defer resp.Body.Close()

    // Read raw response
    data, _ := ioutil.ReadAll(resp.Body)
    err = json.Unmarshal(data, &m)
    if err != nil {
      log.Fatal(err)
    }

    if (!quiet) {
      if (use_color) {
        fmt.Println("[" + magenta("*") + "] Parsing WAF vendors data... (" + cyan(strconv.Itoa(len(strings.Split(string(data), "\n")))) + " lines)")
      } else {
        fmt.Println("[*] Parsing WAF vendors data... (" + strconv.Itoa(len(strings.Split(string(data), "\n"))) + " lines)")
      }
    }
    fmt.Println()

    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
      line := scanner.Text()

      if (line == "") {
        continue
      }

      /*if (!strings.HasPrefix(line, "http://")) && (!strings.HasPrefix(line, "https://")) {
        line = "https://" + line
      }

      if (strings.HasSuffix(line, "=")) {
        line = line + payload

      } else if (!strings.HasSuffix(line, "/")) {
        line = line + "/" + payload
      }*/

      urls_c <- line
    }

    close(urls_c)
    wg.Wait()

  }

  if (!quiet) {
    if (use_color) {
      fmt.Println("\n[" + green("+") + "] Elapsed time:", green(core.TimerDiff(t1)))
    } else {
      fmt.Println("\n[+] Elapsed time:", core.TimerDiff(t1))
    }
  }
}

func Find(o interface{}, key string) (interface{}, bool) { // Function used to find key values
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
      if res, ok := Find(m, key); ok {
        return res, true
      }
    }

    // if the value is an array, search recursively 
    // from each element
    if va, ok := v.([]interface{}); ok {
      for _, a := range va {
        if res, ok := Find(a, key); ok {
          return res, true
        }
      }
    }
  }

  // element not found
  return nil, false
}



