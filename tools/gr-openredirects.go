package main

import (
  "os"
  "fmt"
  "log"
  "net"
  "sync"
  "time"
  "flag"
  "bufio"
  "strings"
  "strconv"
  "net/http"
  "crypto/tls"
  nurl "net/url"
  "github.com/fatih/color"
  tld "github.com/jpillora/go-tld"

  "github.com/D3Ext/go-recon/core"
)

var red func(a ...interface{}) string = color.New(color.FgRed).SprintFunc()
var cyan func(a ...interface{}) string = color.New(color.FgCyan).SprintFunc()
var green func(a ...interface{}) string = color.New(color.FgGreen).SprintFunc()
var magenta func(a ...interface{}) string = color.New(color.FgMagenta).SprintFunc()
var yellow func(a ...interface{}) string = color.New(color.FgYellow).SprintFunc()

func helpPanel(){
  fmt.Println(`Usage of gr-openredirects:
    -u)       url to check open redirect (i.e. https://example.com/?foo=FUZZ)
    -l)       file containing a list of urls to check open redirects (one url per line)
    -k)       keyword to replace in urls with payloads (default=FUZZ)
    -p)       file containing a list of payloads (if not especified, default payloads are used)
    -w)       number of concurrent workers (default=15)
    -m)       requests method (GET, POST, PUT...)
    --proxy)  proxy to send requests through (i.e. http://127.0.0.1:8080)
    -a)       user agent to include on requests (default=none)
    -t)       milliseconds to wait before each request timeout (default=5000)
    -o)       file to write vulnerable urls into
    -c)       use color on output
    -q)       don't print banner nor logging, only output
    -h)       print help panel

Examples:
    gr-openredirects -u https://example.com/?foo=FUZZ -c
    gr-openredirects -u https://example.com/?foo=TEST -k TEST
    gr-openredirects -l urls.txt -p payloads.txt -o vulnerable.txt
    cat urls.txt | gr-openredirects
    `)
}

func main(){
  var url string
  var list string
  var keyword string
  var payloads_list string
  var workers int
  var method string
  var proxy string
  var timeout int
  var user_agent string
  var output string
  var use_color bool
  var quiet bool
  var help bool
  var stdin bool

  flag.StringVar(&url, "u", "", "url to check open redirect (i.e. https://example.com/?foo=FUZZ)")
  flag.StringVar(&list, "l", "", "file containing a list of urls to check open redirects (one url per line)")
  flag.StringVar(&keyword, "k", "FUZZ", "keyword to replace in urls with payloads")
  flag.StringVar(&payloads_list, "p", "", "file containing a list of payloads (if not especified, default payloads are used)")
  flag.IntVar(&workers, "w", 10, "number of concurrent workers")
  flag.StringVar(&method, "m", "GET", "requests method (GET, POST, PUT...)")
  flag.StringVar(&proxy, "proxy", "", "proxy to send requests through (i.e. http://127.0.0.1:8080)")
  flag.IntVar(&timeout, "t", 4000, "milliseconds to wait before reach request timeout")
  flag.StringVar(&output, "o", "", "save vulnerable urls to file")
  flag.StringVar(&user_agent, "a", "", "user agent to include on requests (default=none)")
  flag.BoolVar(&use_color, "c", false, "use color on output")
  flag.BoolVar(&quiet, "q", false, "don't print banner nor logging, only output")
  flag.BoolVar(&help, "h", false, "print help panel")
  flag.Parse()

  t1 := core.StartTimer()

  if (!quiet) {
    fmt.Print(core.Banner())
  }

  if (help) {
    fmt.Println()
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

  var out_f *os.File
  if (output != "") {
    out_f, err = os.Create(output)
    if err != nil {
      log.Fatal(err)
    }
  }

  // Create requests client
  t := time.Duration(timeout) * time.Millisecond

  proxy_url, err := nurl.Parse(proxy)
  if err != nil {
    log.Fatal(err)
  }

  var transport = &http.Transport{}
  if (proxy != "") {
    transport = &http.Transport{
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

  } else {
    transport = &http.Transport{
      MaxIdleConns:      30,
      IdleConnTimeout:   time.Second,
      DisableKeepAlives: true,
      TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, // Disable ssl verify
      DialContext: (&net.Dialer{
        Timeout:   t,
        KeepAlive: time.Second,
      }).DialContext,
    }
  }

  client := &http.Client{ // Create requests client
    Transport:     transport,
    Timeout:       t,
  }

  var payloads []string
  if (payloads_list == "") {
    payloads = core.GetPayloads()

  } else {
    pf, err := os.Open(payloads_list) // Get payloads from given file
    if err != nil {
      log.Fatal(err)
    }
    defer pf.Close()

    sc := bufio.NewScanner(pf)
    for sc.Scan() {
      payloads = append(payloads, sc.Text())
    }
  }

  var redirectTarget string
  for _, p := range payloads { // get payloads redirect domain (i.e. bing.com, example.com, google.com)
    if (strings.HasPrefix(p, "http://")) || (strings.HasPrefix(p, "https://")) {
      parse, err := tld.Parse(p)
      if err != nil {
        continue
      }

      redirectTarget = parse.Domain + "." + parse.TLD
      break
    }
  }

  if (!quiet) {
    fmt.Println()
    core.Magenta("Checking open redirects with " + strconv.Itoa(len(payloads)) + " payloads\n", use_color)
  }

  var counter int
  urls_c := make(chan string) // Create channel which holds urls to use
  payloads_c := make(chan string)
  var wg sync.WaitGroup

  if (url != "") {

    _, err = http.Get(url)
    if err != nil {
      log.Fatal(err)
    }

    for i := 0; i < workers; i++ {
      wg.Add(1)

      go func(){
        for payload := range payloads_c { // test all payloads on current url
          new_url := strings.Replace(url, keyword, payload, -1) // Replace keyword in url with each payload
          req, err := http.NewRequest(method, new_url, nil)
          if err != nil {
            continue
          }

          req.Header.Add("Connection", "close")
          if (user_agent != "") { // Check if user agent has value
            req.Header.Set("User-Agent", user_agent)
          }
          req.Close = true

          resp, err := client.Do(req) // Send requests with out custom client config
          if err != nil {
            continue
          }
          defer resp.Body.Close()

          if resp.StatusCode == http.StatusOK {
            if strings.Contains(resp.Request.URL.String(), redirectTarget) {
              fmt.Println(new_url)
              counter += 1

              if (output != "") { // Write url with payload to output file
                _, err = out_f.WriteString(new_url + "\n")
                if err != nil {
                  log.Fatal(err)
                }
              }
            }
          }
        }

        wg.Done()
      }()
    }

    for _, p := range payloads {
      payloads_c <- p
    }

    close(payloads_c)
    wg.Wait()

  } else if (list != "") || (stdin) {

    for i := 0; i < workers; i++ { // create n workers
      wg.Add(1)

      go func(){
        for u := range urls_c { // iterate over urls
          current_url := u

          for _, payload := range payloads { // test all payloads current url
            new_url := strings.Replace(current_url, keyword, payload, -1) // Replace keyword in url with each payload
            req, err := http.NewRequest(method, new_url, nil)
            if err != nil {
              continue
            }

            req.Header.Add("Connection", "close")
            if (user_agent != "") { // Check if user agent has value
              req.Header.Set("User-Agent", user_agent)
            }
            req.Close = true

            resp, err := client.Do(req) // Send requests with out custom client config
            if err != nil {
              continue
            }
            defer resp.Body.Close()

            if resp.StatusCode == http.StatusOK {
              if strings.Contains(resp.Request.URL.String(), redirectTarget) {
                fmt.Println(new_url)
                counter += 1

                if (output != "") { // Write url with payload to output file
                  _, err = out_f.WriteString(new_url + "\n")
                  if err != nil {
                    log.Fatal(err)
                  }
                }
              }
            }

          }

        }

        wg.Done() // Finish worker
      }()
    }

    var f *os.File
    var err error

    // Get file descriptor from file or stdin
    if (list != "") {
      f, err = os.Open(list)
      if err != nil {
        log.Fatal(err)
      }
      defer f.Close()

    } else if (stdin) {
      f = os.Stdin
    }

    scanner := bufio.NewScanner(f)
    for scanner.Scan() { // Iterate over every single line
      line := scanner.Text()

      if (line != "") {
        urls_c <- line
      }
    }

    close(urls_c)   // Close channel
    wg.Wait()  // and wait for urls workers wait group

  }

  // Finally some logging to aid users
  if (!quiet) {
    if (counter >= 1) {
      fmt.Println()
      core.Green(strconv.Itoa(counter) + " open redirects found", use_color)
    } else {
      core.Red("No open redirect found!", use_color)
    }

    if (output != "") {
      if (counter >= 1) { // Check if at least one url was vulnerable to open redirect
        if (use_color) {
          fmt.Println("[" + green("+") + "] Urls written to", output)
        } else {
          fmt.Println("[+] Urls written to", output)
        }
      }
    }

    if (use_color) {
      if (output != "" || counter >= 1) {
        fmt.Println("[" + green("+") + "] Elapsed time:", green(core.TimerDiff(t1)))
      } else {
        fmt.Println("\n[" + green("+") + "] Elapsed time:", green(core.TimerDiff(t1)))
      }
    } else {
      if (output != "" || counter >= 1) {
        fmt.Println("[+] Elapsed time:", core.TimerDiff(t1))
      } else {
        fmt.Println("\n[+] Elapsed time:", core.TimerDiff(t1))
      }
    }
  }
}



