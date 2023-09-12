package main

import (
  "os"
  "fmt"
  "log"
  "flag"
  "bufio"
  "strings"
  "net/url"
  "encoding/json"
  "github.com/fatih/color"

  "github.com/D3Ext/go-recon/core"
)

var red func(a ...interface{}) string = color.New(color.FgRed).SprintFunc()
var cyan func(a ...interface{}) string = color.New(color.FgCyan).SprintFunc()
var green func(a ...interface{}) string = color.New(color.FgGreen).SprintFunc()
var magenta func(a ...interface{}) string = color.New(color.FgMagenta).SprintFunc()

type UrlsInfo struct {
  Urls []string   `json:"urls"`
  Length int      `json:"length"`
  Time string     `json:"time"`
}

func helpPanel() {
  fmt.Println(`Usage of gr-replace:
    -l)       file containing a list of urls to replace specific strings on them (one url per line)
    -k)       keyword to replace in urls with supplied value (i.e. FUZZ)
    -p)       parameter name to replace it value with supplied value (i.e. id)
    -s)       value to replace keyword with in urls (i.e. <script>alert('XSS')</script> )
    --hide)   don't print urls that aren't modified
    -o)       file to write modified urls into
    -oj)      file to write modified urls into (JSON format)
    -c)       print colors on output
    -q)       don't print banner nor logging, only output
    -h)       print help panel

Examples:
    gr-replace -l urls.txt -k FUZZ -s "<script>alert('XSS')</script>" -o new_urls.txt -q
    gr-replace -l urls.txt -p id -s "' or 1=1-- -" -c
    cat urls.txt | gr-replace --hide
    `)
}

func main(){
  var list string
  var keyword string
  var param string
  var value string
  var hide bool
  var output string
  var json_output string
  var stdin bool
  var quiet bool
  var use_color bool
  var help bool

  flag.StringVar(&list, "l", "", "file containing a list of urls to replace specific strings on them (one url per line)")
  flag.StringVar(&keyword, "k", "", "keyword to replace in urls with supplied value (i.e. FUZZ)")
  flag.StringVar(&param, "p", "", "parameter name to replace it value with supplied value (i.e. id)")
  flag.StringVar(&value, "s", "", "value to replace keyword with in urls (i.e. )")
  flag.BoolVar(&hide, "hide", false, "don't print urls that aren't modified")
  flag.StringVar(&output, "o", "", "file to write modified urls into")
  flag.StringVar(&json_output, "oj", "", "file to write modified urls into (JSON format)")
  flag.BoolVar(&quiet, "q", false, "don't print banner, only output")
  flag.BoolVar(&use_color, "c", false, "print colors on output")
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

  var err error

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

  if (list == "") && (!stdin) { // check if required arguments were given
    helpPanel()
    os.Exit(0)
  }

  if (value == "") || (param == "" && keyword == "") || (param != "" && keyword != "") {
    helpPanel()
    core.Red("You need to provide a valid keyword (-k) or param value (-p) to replace with supplied value (-s)", use_color)
    os.Exit(0)
  }

  var out_f *os.File
  if (output != "") { // create output file if it was provided
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

  var counter int
  var replaced_urls []string

  if (!quiet) {
    core.Magenta("Replacing values...\n", use_color)
  }

  if (list != "") || (stdin) {

    var f *os.File

    if list != "" {
      f, err = os.Open(list)
      if err != nil {
        log.Fatal(err)
      }
      defer f.Close()

    } else {
      f = os.Stdin
    }

    var line string
    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
      line_pre_change := scanner.Text()

      u, _ := url.Parse(line_pre_change) // parse each url
      if u == nil {
        continue
      }

      if (keyword != "") && (value != "") {
        line = strings.ReplaceAll(line_pre_change, keyword, value)

      } else if (param != "") && (value != "") {
        for k, _ := range u.Query() { // iterate over parameters and their values
          if (k == param) { // check param name
            line = strings.ReplaceAll(line_pre_change, u.Query().Get(k), value) // replace param value with provided value to replace it
            break
          } else {
            line = line_pre_change
          }
        }
      }

      if (line_pre_change != line) && (hide) {
        fmt.Println(line)
        replaced_urls = append(replaced_urls, line)
        counter += 1

      } else if (!hide) {

        if line != "" {
          fmt.Println(line)
          replaced_urls = append(replaced_urls, line)
        } else {
          fmt.Println(line_pre_change)
          replaced_urls = append(replaced_urls, line_pre_change)
        }

        counter += 1
      }

      if (output != "") {
        _, err = out_f.WriteString(line + "\n")
        if err != nil {
          log.Fatal(err)
        }
      }
    }
  }

  if (json_output != "") {
    json_urls := UrlsInfo{
      Urls: replaced_urls,
      Length: len(replaced_urls),
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

  // finally add some logging to aid users
  if (!quiet) {
    if (counter >= 1) {
      fmt.Println()
      if (output != "") {
        core.Green("Results written to " + output, use_color)
      } else if (json_output != "") {
        core.Green("Results written to " + json_output, use_color)
      }

      if (use_color) {
        fmt.Println("[" + green("+") + "]", cyan(counter), "lines processed")
        fmt.Println("[" + green("+") + "] Elapsed time:", green(core.TimerDiff(t1)))
      } else {
        fmt.Println("[+]", counter, "lines processed")
        fmt.Println("[+] Elapsed time:", core.TimerDiff(t1))
      }

    } else {
      if (use_color) {
        fmt.Println("\n[" + green("+") + "] Elapsed time:", green(core.TimerDiff(t1)))
      } else {
        fmt.Println("\n[+] Elapsed time:", core.TimerDiff(t1))
      }
    }
  }
}


