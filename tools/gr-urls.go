package main

import (
  "io"
  "os"
  "log"
  "fmt"
  "sync"
  "flag"
  "strings"
  "strconv"
  "net/http"
  "io/ioutil"
  "encoding/json"
  "github.com/fatih/color"

  "github.com/D3Ext/go-recon/core"
)

var red func(a ...interface{}) string = color.New(color.FgRed).SprintFunc()
var cyan func(a ...interface{}) string = color.New(color.FgCyan).SprintFunc()
var green func(a ...interface{}) string = color.New(color.FgGreen).SprintFunc()
var magenta func(a ...interface{}) string = color.New(color.FgMagenta).SprintFunc()
var yellow func(a ...interface{}) string = color.New(color.FgYellow).SprintFunc()

type UrlsInfo struct {
  Urls []string `json:"urls"`
  Length int    `json:"length"`
  Time string   `json:"time"`
}

func helpPanel() {
  fmt.Println(`Usage of gr-urls:
    -d)       domain to retrieve a bunch of urls from different sources (i.e. example.com)
    -w)       number of concurrent workers (default=2)
    -o)       file to write urls into
    -oj)      file to write urls into (JSON format)
    -r)       retrieve urls just for given domain without subdomains (default=disabled)
    -t)       milliseconds to wait before each request timeout (default=15000)
    -c)       print colors on output
    -q)       don't print banner, only output
    -h)       print help panel

Examples:
    gr-urls -d example.com -o urls.txt
    gr-urls -d example.com -r
    echo "example.com" | gr-urls
    `)
}

var found_urls []string
var current_url string
var counter int

func main(){
  var domain string
  var workers int
  var output string
  var json_output string
  var recursive bool
  var stdin bool
  var timeout int
  var quiet bool
  var use_color bool
  var help bool

  flag.StringVar(&domain, "d", "", "domain to retrieve a bunch of urls from different sources (i.e. example.com)")
  flag.IntVar(&workers, "w", 2, "number of concurrent workers")
  flag.StringVar(&output, "o", "", "file to write urls into")
  flag.StringVar(&json_output, "oj", "", "file to write urls into (JSON format)")
  flag.BoolVar(&recursive, "r", false, "retrieve urls just for given domain without subdomains (default=disabled)")
  flag.IntVar(&timeout, "t", 15000, "milliseconds to wait before each request timeout")
  flag.BoolVar(&quiet, "q", false, "don't print banner, only output")
  flag.BoolVar(&use_color, "c", false, "print colors on output")
  flag.BoolVar(&help, "h", false, "print help panel")
  flag.Parse()

  t1 := core.StartTimer()

  if (!quiet) {
    fmt.Println(core.Banner())
  }

  if (help) { // Print help panel
    helpPanel()
    os.Exit(0)
  }

  fi, err := os.Stdin.Stat()
  if err != nil {
    log.Fatal(err)
  }

  if fi.Mode() & os.ModeNamedPipe == 0 {
    stdin = false // stdin is empty
  } else {
    stdin = true // stdin has value
  }

  if (stdin) {
    // Read STDIN
    stdin, err := io.ReadAll(os.Stdin)
    if err != nil {
      log.Fatal(err)
    }

    if (string(stdin) != "") {
      domain = strings.TrimSuffix(string(stdin), "\n")
    }
  }

  if (domain == "") && (!stdin) {
    helpPanel()
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

  if (!quiet) {
    core.Magenta("Gathering urls...\n", use_color)
  }

  if (recursive) {
    recursive = false
  } else {
    recursive = true
  }

  //results := make(chan string) // urls will be sent through this channel
  var wg sync.WaitGroup

  client := core.CreateHttpClient(timeout)

  var max_page int
  var page_url string

  if (!recursive) {
    page_url = "http://web.archive.org/cdx/search/cdx?url=" + domain + "&showNumPages=true"
  } else {
    page_url = "http://web.archive.org/cdx/search/cdx?url=*." + domain + "&showNumPages=true"
  }

  req, err := http.NewRequest("GET", page_url, nil)
  if err != nil {
    core.Red("WaybackMachine seems to be slow or down, try again", use_color)
    log.Fatal(err)
  }

  res, err := client.Do(req)
  if err != nil {
    core.Red("WaybackMachine seems to be slow or down, try again", use_color)
    log.Fatal(err)
  }

  raw, err := ioutil.ReadAll(res.Body)
  if err != nil {
  }

  max_page, err = strconv.Atoi(strings.TrimSuffix(string(raw), "\n"))
  if err != nil {
    log.Fatal(err)
  }

  var rangeSize int // split number of pages between workers to get the total number of iterations

  if (max_page % workers == 0) { // check if workers is multiple of max_page
    rangeSize = max_page / workers

  } else if (max_page % workers != 0) && (workers > 1) { // if not, subtract 1 or add 5 as maximum to get a 
    for i := -1; i <= 5; i++ {
      if (max_page % (workers+i) == 0) {
        workers = workers + i
        break
      }
    }

    rangeSize = max_page / workers

  } else if (max_page >= 1) && (max_page <= 6) {
    rangeSize = 10
    workers = 1
  }

  for i := 0; i < workers; i++ { // create n workers
    wg.Add(1)

    go worker(domain, client, recursive, rangeSize*i, rangeSize*(i+1), &wg, output, out_f)
  }

  wg.Wait()

  if (json_output != "") {
    json_urls := UrlsInfo{
      Urls: found_urls,
      Length: len(found_urls),
      Time: core.TimerDiff(t1).String(),
    }

    json_body, err := json.Marshal(json_urls)
    if err != nil {
      log.Fatal(err)
    }

    _, err = out_f.WriteString(string(json_body))
    if err != nil {
      log.Fatal(err)
    }
  }

  if (!quiet) {
    if (output != "" || json_output != "") && (counter >= 1) {
      fmt.Println()
      if (output != "") {
        core.Green("Urls written to " + output, use_color)
      } else if (json_output != "") {
        core.Green("Urls written to " + json_output, use_color)
      }
      if (use_color) {
        fmt.Println("[" + green("+") + "]", cyan(counter), "urls obtained")
      } else {
        fmt.Println("[+]", counter, "urls obtained")
      }

    } else if (output == "") && (json_output == "") {
      if (use_color) {
        fmt.Println("\n[" + green("+") + "]", cyan(counter), "urls obtained")
      } else {
        fmt.Println("\n[+]", counter, "urls obtained")
      }
    }

    if (use_color) {
      if (output != "" || json_output != "" || counter >= 1) {
        fmt.Println("[" + green("+") + "] Elapsed time:", green(strings.ReplaceAll(core.TimerDiff(t1).String(), "m", "m ")))
      } else {
        fmt.Println("\n[" + green("+") + "] Elapsed time:", green(strings.ReplaceAll(core.TimerDiff(t1).String(), "m", "m ")))
      }
    } else {
      if (output != "" || json_output != "" || counter >= 1) {
        fmt.Println("[+] Elapsed time:", strings.ReplaceAll(core.TimerDiff(t1).String(), "m", "m "))
      } else {
        fmt.Println("\n[+] Elapsed time:", strings.ReplaceAll(core.TimerDiff(t1).String(), "m", "m "))
      }
    }
  }
}

func worker(domain string, client *http.Client, recursive bool, start int, end int, wg *sync.WaitGroup, output string, out_f *os.File) {
  var err_counter int

  for i := start; i < end; i++ {

    if (err_counter >= 5) {
      break
    }

    if (!recursive) {
      current_url = "http://web.archive.org/cdx/search/cdx?url=" + domain + "/*&output=json&collapse=urlkey&page=" + strconv.Itoa(i)
    } else {
      current_url = "http://web.archive.org/cdx/search/cdx?url=*." + domain + "/*&output=json&collapse=urlkey&page=" + strconv.Itoa(i)
    }
    //fmt.Println(current_url)

    req, err := http.NewRequest("GET", current_url, nil)
    if err != nil {
      err_counter += 1
      continue
    }

    res, err := client.Do(req)
    if err != nil {
      err_counter += 1
      continue
    }
    defer res.Body.Close()

    raw, err := ioutil.ReadAll(res.Body)
    if err != nil {
      err_counter += 1
      continue
    }

    if (err_counter >= 5) {
      break
    }

    var wrapper [][]string
    err = json.Unmarshal(raw, &wrapper)
    if err != nil {
      err_counter += 1
      continue
    }

    if (err_counter >= 5) {
      break
    }

    skip := true
    for _, urls := range wrapper {
      if skip {
        skip = false
        continue
      }

      fmt.Println(urls[2])
      found_urls = append(found_urls, urls[2])
      counter += 1
      if (output != "") {
        _, err = out_f.WriteString(urls[2] + "\n")
        if err != nil {
          log.Fatal(err)
        }
      }
    }

    //fmt.Println(current_url, "- succeed!")
  }

  wg.Done()
}


