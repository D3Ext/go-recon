package main

import (
  "os"
  "fmt"
  "log"
  "net"
  "time"
  "flag"
  "sync"
  "bufio"
  "strings"
  "net/http"
  "io/ioutil"
  "crypto/tls"
  "github.com/fatih/color"
  wappalyzer "github.com/projectdiscovery/wappalyzergo"

  "github.com/D3Ext/go-recon/core"
)

var red func(a ...interface{}) string = color.New(color.FgRed).SprintFunc()
var cyan func(a ...interface{}) string = color.New(color.FgCyan).SprintFunc()
var green func(a ...interface{}) string = color.New(color.FgGreen).SprintFunc()
var magenta func(a ...interface{}) string = color.New(color.FgMagenta).SprintFunc()
var yellow func(a ...interface{}) string = color.New(color.FgYellow).SprintFunc()

func helpPanel() {
  fmt.Println(`Usage of gr-tech:
    -u)       url to identify its technologies (i.e. https://example.com)
    -l)       file containing a list of urls to identify their technologies (one url per line)
    -w)       number of concurrent workers (default=15)
    -a)       user agent to include on requests (default=none)
    -c)       print colors on output
    -t)       milliseconds to wait before each request timeout (default=4000)
    -q)       don't print banner nor logging, only output
    -h)       print help panel
  
Examples:
    gr-tech -u https://example.com -w 10
    gr-tech -l urls.txt -c
    cat urls.txt | gr-tech -q
    `)
}

func main(){
  var url string
  var list string
  var workers int
  var timeout int
  var user_agent string
  var use_color bool
  var quiet bool
  var help bool
  var stdin bool

  flag.StringVar(&url, "u", "", "url to identify its technologies (i.e. https://example.com)")
  flag.StringVar(&list, "l", "", "list of urls to identify their technologies (one domain per line)")
  flag.IntVar(&workers, "w", 15, "number of concurrent workers")
  flag.IntVar(&timeout, "t", 4000, "milliseconds to wait before each request timeout")
  flag.StringVar(&user_agent, "a", "", "user agent to include on requests (default=none)")
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

  // Create requests client
  t := time.Duration(timeout * 1000000)

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

  if (!quiet) {
    core.Magenta("Identifying technologies running on targets...\n", use_color)
  }

  var counter int
  if (url != "") {
    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Add("Connection", "close")
    if (user_agent != "") {
      req.Header.Set("User-Agent", user_agent)
    }
    req.Close = true

    resp, err := client.Do(req) // Send request
    if err != nil {
      log.Fatal(err)
    } else {
      data, err := ioutil.ReadAll(resp.Body)
      if err != nil {
        log.Fatal(err)
      }

      wappalyzer_client, err := wappalyzer.New()
      if err != nil {
        log.Fatal(err)
      }

      techs := wappalyzer_client.Fingerprint(resp.Header, data)
      var techs_array []string

      if len(techs) != 0 {
        fmt.Printf(url + " - [")
      } else {
        fmt.Print(url)
      }
      for k, _ := range techs {
        techs_array = append(techs_array, k)
      }

      for _, v := range techs_array {
        if (techs_array[len(techs_array)-1] == v) {
          fmt.Print(v + "]")
        } else {
          fmt.Print(v + ", ")
        }
      }
      fmt.Println()
    }

  } else if (list != "") || (stdin) {

    var f *os.File
    var err error

    if (list != "") {
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

    for i := 0; i < workers; i++ {
      wg.Add(1)

      go func(){
        for u := range urls_c {
          counter += 1
          req, _ := http.NewRequest("GET", u, nil)
          req.Header.Add("Connection", "close")
          if (user_agent != "") {
            req.Header.Set("User-Agent", user_agent)
          }
          req.Close = true

          resp, err := client.Do(req) // Send request
          if err != nil {
            log.Fatal(err)
          } else {
            data, err := ioutil.ReadAll(resp.Body)
            if err != nil {
              log.Fatal(err)
            }

            wappalyzer_client, err := wappalyzer.New()
            if err != nil {
              log.Fatal(err)
            }

            techs := wappalyzer_client.Fingerprint(resp.Header, data)
            var techs_array []string

            if len(techs) != 0 {
              fmt.Printf(u + " - [")
            } else {
              fmt.Print(u)
            }

            for k, _ := range techs {
              techs_array = append(techs_array, k)
            }

            for _, v := range techs_array {
              if (techs_array[len(techs_array)-1] == v) {
                if (use_color) {
                  fmt.Print(cyan(v) + "]")
                } else {
                  fmt.Print(v + "]")
                }
              } else {
                if (use_color) {
                  fmt.Print(cyan(v) + ", ")
                } else {
                  fmt.Print(v + ", ")
                }
              }
            }
            fmt.Println()
          }
        }

        wg.Done()
      }()
    }

    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
      line := scanner.Text()

      if (line == "") {
        continue
      }

      if (!strings.HasPrefix(line, "http://")) && (!strings.HasPrefix(line, "https://")) {
        line = "https://" + line
      }

      urls_c <- line
    }

    close(urls_c)
    wg.Wait()
  }

  if (!quiet) {
    if (counter > 1) {
      if (use_color) {
        fmt.Println("\n[" + green("+") + "]", counter, "urls were processed")
      } else {
        fmt.Println("\n[+]", counter, "urls were processed")
      }
    }

    if (use_color) {
      if (counter > 1) {
        fmt.Println("[" + green("+") + "] Elapsed time:", green(core.TimerDiff(t1)))
      } else {
        fmt.Println("\n[" + green("+") + "] Elapsed time:", green(core.TimerDiff(t1)))
      }
    } else {
      if (counter >= 1) {
        fmt.Println("[+] Elapsed time:", core.TimerDiff(t1))
      } else {
        fmt.Println("\n[+] Elapsed time:", core.TimerDiff(t1))
      }
    }
  }
}

