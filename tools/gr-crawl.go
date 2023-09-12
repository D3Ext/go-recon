package main

import (
  "os"
  "fmt"
  "log"
  "flag"
  "time"
  "bufio"
  "regexp"
  "strings"
  "net/http"
  "crypto/tls"
  nurl "net/url"
  "encoding/json"
  "github.com/fatih/color"
  "github.com/gocolly/colly/v2"

  "github.com/D3Ext/go-recon/core"
)

var red func(a ...interface{}) string = color.New(color.FgRed).SprintFunc()
var cyan func(a ...interface{}) string = color.New(color.FgCyan).SprintFunc()
var green func(a ...interface{}) string = color.New(color.FgGreen).SprintFunc()
var magenta func(a ...interface{}) string = color.New(color.FgMagenta).SprintFunc()

type CrawlInfo struct {
  Urls []string   `json:"urls"`
  Length int      `json:"length"`
  Time string     `json:"time"`
}

func helpPanel() {
  fmt.Println(`Usage of gr-crawl:
    -u)       target url to crawl to gather urls (i.e. https://example.com)
    -l)       file containing a list of urls to crawl (one per line)
    -d)       depth to crawl (default=2)
    -w)       number of concurrent workers (default=10)
    -p)       proxy to send requests through (i.e. http://127.0.0.1:8080)
    -o)       file to write urls into
    -oj)      file to write urls into (JSON format)
    -s)       also crawl subdomains (default=disabled)
    -c)       print colors on output
    -t)       millisecond to wait before each request timeout (default=5000)
    -q)       don't print banner nor logging, only output
    -h)       print help panel
  
Examples:
    gr-crawl -u https://example.com -o endpoints.txt -c
    gr-crawl -l urls.txt -d 3 -w 15
    cat urls.txt | gr-crawl -q
    `)
}

func main(){
  var url string
  var list string
  var depth int
  var workers int
  var proxy string
  var output string
  var json_output string
  var subs bool
  var use_color bool
  var timeout int
  var quiet bool
  var help bool
  var stdin bool

  flag.StringVar(&url, "u", "", "target url to crawl to gather urls (i.e. https://example.com)")
  flag.StringVar(&list, "l", "", "file containing a list of urls to crawl (one per line)")
  flag.IntVar(&depth, "d", 2, "depth to crawl")
  flag.IntVar(&workers, "w", 10, "number of concurrent workers")
  flag.StringVar(&proxy, "p", "", "proxy to send requests through (i.e. http://127.0.0.1:8080)")
  flag.StringVar(&output, "o", "", "file to write urls into")
  flag.StringVar(&json_output, "oj", "", "file to write urls into (JSON format)")
  flag.BoolVar(&subs, "s", false, "also crawl subdomains (default=disabled)")
  flag.BoolVar(&quiet, "q", false, "don't print banner, only output")
  flag.BoolVar(&use_color, "c", false, "print colors on output")
  flag.IntVar(&timeout, "t", 5000, "millisecond to wait before each request timeout (default=5000)")
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
    stdin = false
  } else {
    stdin = true
  }

  // if url, list and stdin parameters are empty print help panel and exit
  if (url == "") && (list == "") && (!stdin) {
    helpPanel()
    core.Red("Especify a valid argument (-u) or (-l)", use_color)
    os.Exit(0)
  }

  var out_f *os.File
  if (output != "") {
    out_f, err = os.Create(output)
    if err != nil {
      log.Fatal(err)
    }
  } else if (json_output != "") {
    out_f, err = os.Create(json_output)
    if err != nil {
      log.Fatal(err)
    }
  }

  results := make(chan string)

  var counter int
  var urls []string

  // add url or urls to array
  if (url != "") {
    urls = append(urls, url)

  } else if (list != "") {
    f, err := os.Open(list)
    if err != nil {
      log.Fatal(err)
    }
    defer f.Close()

    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
      urls = append(urls, scanner.Text())
    }

  } else if (stdin) {
    f := os.Stdin

    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
      urls = append(urls, scanner.Text())
    }
  }

  go func() {
    for _, url := range urls {
      url_parse, err := nurl.Parse(url)
      if err != nil {
        log.Fatal(err)
      }

      hostname := url_parse.Hostname()
      allowed_doms := []string{hostname}

      c := colly.NewCollector(
        colly.UserAgent("Mozilla/5.0 (X11; Linux x86_64; rv:78.0) Gecko/20100101 Firefox/78.0"),
        colly.AllowedDomains(allowed_doms...),
        colly.MaxDepth(depth),
        colly.Async(true),
      )

      if subs {
        c.AllowedDomains = nil
        c.URLFilters = []*regexp.Regexp{regexp.MustCompile(".*(\\.|\\/\\/)" + strings.ReplaceAll(hostname, ".", "\\.") + "((#|\\/|\\?).*)?")}
      }

      c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: workers})

      c.OnHTML("a[href]", func(e *colly.HTMLElement) {
        link := e.Attr("href")
        abs_link := e.Request.AbsoluteURL(link)
        if strings.Contains(abs_link, url) {
          printResult(link, "href", results, e)
          e.Request.Visit(link)
        }
      })

      c.OnHTML("script[src]", func(e *colly.HTMLElement) {
        printResult(e.Attr("src"), "script", results, e)
      })

      c.OnHTML("form[action]", func(e *colly.HTMLElement) {
        printResult(e.Attr("action"), "form", results, e)
      })

      c.WithTransport(&http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
      })

      finished := make(chan int, 1)

      go func() {
        c.Visit(url)
        c.Wait()

        finished <- 0
      }()

      select {
      case _ = <-finished:
        close(finished)
        continue
      case <-time.After(time.Duration(timeout) * time.Millisecond):
        continue
      }
    }

    close(results)
  }()

  var unique_urls []string

  for res := range results { // receive urls from channel

    found := 0
    for _, u := range unique_urls { // iterate over unique urls array
      if (res == u) { // if current url is present in array, skip this url
        found = 1
        break
      }
    }

    if (found == 1) {
      continue
    }

    unique_urls = append(unique_urls, res) // append url to unique urls array

    counter += 1
    fmt.Println(res)

    if (output != "") {
      _, err = out_f.WriteString(res + "\n")
      if err != nil {
        log.Fatal(err)
      }
    }
  }

  if (json_output != "") {
    json_crawl := CrawlInfo{
      Urls: unique_urls,
      Length: len(unique_urls),
      Time: core.TimerDiff(t1).String(),
    }

    json_body, err := json.Marshal(json_crawl)
    if err != nil {
      log.Fatal(err)
    }

    _, err = out_f.WriteString(string(json_body))
    if err != nil {
      log.Fatal(err)
    }
  }

  if (!quiet) {
    if (counter >= 1) {
      fmt.Println()
      if (output != "") {
        core.Green("Urls written to " + output, use_color)
      } else if (json_output != "") {
        core.Green("Urls written to " + json_output, use_color)
      }

      if (use_color) {
        fmt.Println("[" + green("+") + "]", counter, "urls found")
      } else {
        fmt.Println("[+]", counter, "urls found")
      }
    }

    if (use_color) {
      if (counter >= 1) {
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


func printResult(link string, sourceName string, results chan string, e *colly.HTMLElement) {
	result := e.Request.AbsoluteURL(link)
	if result != "" {
		// If timeout occurs before goroutines are finished, recover from panic that may occur when attempting writing 
    // to results to closed results channel
		defer func() {
			if err := recover(); err != nil {
				return
			}
		}()

    results <- result
	}
}



