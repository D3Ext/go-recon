package main

import (
  "os"
  "fmt"
  "log"
  "flag"
  "time"
  "bufio"
  "strings"
  "encoding/json"
  "github.com/fatih/color"

  "github.com/D3Ext/go-recon/core"
)

var red func(a ...interface{}) string = color.New(color.FgRed).SprintFunc()
var cyan func(a ...interface{}) string = color.New(color.FgCyan).SprintFunc()
var green func(a ...interface{}) string = color.New(color.FgGreen).SprintFunc()
var magenta func(a ...interface{}) string = color.New(color.FgMagenta).SprintFunc()
var yellow func(a ...interface{}) string = color.New(color.FgYellow).SprintFunc()

type SubdomainsInfo struct {
  Subdomains []string   `json:"subdomains"`
  Length int            `json:"length"`
  Time string           `json:"time"` 
}

func helpPanel() {
  fmt.Println(`Usage of gr-subdomains:
    -d)       domain to find its subdomains (i.e. example.com)
    -l)       file containing a list of domains to find their subdomains (one domain per line)
    -o)       file to write subdomains into
    -oj)      file to write subdomains into (JSON format)
    -c)       print colors on output
    -t)       milliseconds to wait before each request timeout (default=5000)
    -q)       don't print banner nor logging, only output
    -h)       print help panel
  
Examples:
    gr-subdomains -d example.com -o subdomains.txt -c
    gr-subdomains -l domains.txt
    cat domains.txt | gr-subdomains -q
    `)
}

func main(){
  var domain string
  var list string
  var output string
  var json_output string
  var stdin bool
  var timeout int
  var quiet bool
  var use_color bool
  var help bool

  flag.StringVar(&domain, "d", "", "domain to find its subdomains (i.e. example.com)")
  flag.StringVar(&list, "l", "", "file containing a list of domains to find their subdomains (one domain per line)")
  flag.StringVar(&output, "o", "", "file to write subdomains into")
  flag.StringVar(&json_output, "oj", "", "file to write subdomains into (JSON format)")
  flag.IntVar(&timeout, "t", 5000, "milliseconds to wait before each request timeout")
  flag.BoolVar(&quiet, "q", false, "don't print banner nor logging, only output")
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

  // if domain, list and stdin parameters are empty print help panel and exit
  if (domain == "") && (list == "") && (!stdin) {
    helpPanel()
    os.Exit(0)
  }

  if (domain != "") && (list != "") {
    helpPanel()
    core.Red("You can't use (-d) and (-l) at same time", use_color)
    os.Exit(0)
  }

  // define variables which will be used to write subdomains to file
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

  var counter int
  var total_counter int

  if (domain != "") { // if domain parameter has value enter here

    if (!quiet) {
      core.Magenta("Finding " + domain + " subdomains...\n", use_color)
    }

    // Create channel for asynchronous jobs
    subdoms := make(chan string)
    var subdomains []string

    // Run go routine
    go core.GetAllSubdomains(domain, subdoms, timeout)

    for sub := range subdoms { // iterate over subdomains received through channel
      subdomains = append(subdomains, sub)
    }

    for _, res := range subdomains {
      if (res != "") && (!strings.HasPrefix(res, "*.")) {
        fmt.Println(res)
        counter += 1

        if (output != "") { // if output parameter has value, write subdomains to file
          _, err = out_f.WriteString(res + "\n")
          if err != nil {
            log.Fatal(err)
          }
        }
      }
    }

  if (json_output != "") {
    json_subs := SubdomainsInfo{
      Subdomains: subdomains,
      Length: counter,
      Time: core.TimerDiff(t1).String(),
    }

    json_body, err := json.Marshal(json_subs)
    if err != nil {
      log.Fatal(err)
    }

    _, err = out_f.WriteString(string(json_body))
    if err != nil {
      log.Fatal(err)
    }
  }


  } else if (list != "") || (stdin) { // enter here if list parameter has value or if stdin has value

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

    var doms_length int
    var domains []string

    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
      if (scanner.Text() != "") && (strings.Contains(scanner.Text(), ".")) {
        doms_length += 1
        domains = append(domains, scanner.Text())
      }
    }

    if (!quiet) {
      if (use_color) {
        fmt.Println("[" + magenta("*") + "] Finding subdomains for", green(doms_length), "domains")
      } else {
        fmt.Println("[*] Finding subdomains for", doms_length, "domains")
      }
    }

    for _, dom := range domains {
      if (!quiet) {
        if (use_color) {
          fmt.Println("[" + magenta("*") + "] " + "Domain: " + green(dom))
        } else {
          fmt.Println("[*] Domain: " + dom)
        }
      }

      // Create channel for asynchronous jobs
      subdoms := make(chan string)
      // Run go routine
      go core.GetAllSubdomains(dom, subdoms, timeout)

      for sub := range subdoms { // iterate over subdomains received through channel
        if (sub != "") {
          fmt.Println(sub)
          counter += 1
          total_counter += 1
        }

        if (output != "") { // if output parameter has value, write subdomains to file
          _, err = out_f.WriteString(sub + "\n")
          if err != nil {
            log.Fatal(err)
          }
        }
      }

      // Empty channel
      for len(subdoms) > 0 {
        <-subdoms
      }

      if (!quiet) && (dom != "") {//(dom != domains[len(domains)-1]) {
        if (use_color) {
          fmt.Println("[" + green("+") + "]", cyan(counter), "subdomains found for " + dom + "\n")
        } else {
          fmt.Println("[+]", counter, "subdomains found for " + dom + "\n")
        }
      }

      counter = 0
      time.Sleep(50 * time.Millisecond)
    }
  }

  // Finally some logging to aid users
  if (!quiet) {
    if (total_counter >= 1) {
      if (use_color) {
        fmt.Println("[" + green("+") + "]", cyan(total_counter), "subdomains found in total")
      } else {
        fmt.Println("[+]", total_counter, "subdomains found in total")
      }
    } else if (counter >= 1) {
      if (use_color) {
        fmt.Println("\n[" + green("+") + "]", cyan(counter), "subdomains found")
      } else {
        fmt.Println("\n[+]", counter, "subdomains found")
      }
    }

    if (total_counter >= 1 || counter >= 1) { // Check if at least one url was vulnerable to open redirect
      if (output != "") {
        core.Green("Subdomains written to " + output, use_color)
      } else if (json_output != "") {
        core.Green("Subdomains written to " + json_output, use_color)
      }
    }

    if (use_color) {
      fmt.Println("[" + green("+") + "] Elapsed time:", green(core.TimerDiff(t1)))
    } else {
      fmt.Println("[+] Elapsed time:", core.TimerDiff(t1))
    }
  }
}

