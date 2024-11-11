package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"io/ioutil"
	"log"
	"net/http"
	nu "net/url"
	"os"
  "slices"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/D3Ext/go-recon/core"
)

var red func(a ...interface{}) string = color.New(color.FgRed).SprintFunc()
var cyan func(a ...interface{}) string = color.New(color.FgCyan).SprintFunc()
var blue func(a ...interface{}) string = color.New(color.FgBlue).SprintFunc()
var green func(a ...interface{}) string = color.New(color.FgGreen).SprintFunc()
var magenta func(a ...interface{}) string = color.New(color.FgMagenta).SprintFunc()
var yellow func(a ...interface{}) string = color.New(color.FgYellow).SprintFunc()

type UrlsInfo struct {
	Results   []WafResult `json:"results"`
	Length int            `json:"length"`
	Time   string         `json:"time"`
}

type WafResult struct {
  Url string `json:"url"`
  Waf string `json:"waf"`
}

func helpPanel() {
	fmt.Println(`Usage of gr-waf:
  INPUT:
    -u, -url string       url to identify its WAF (i.e. https://example.com)
    -l, -list string      file containing a list of urls to identify their WAFs (one url per line)

  OUTPUT:
    -o, -output string          file to write discovered WAFs into (TXT format)
    -oj, -output-json string    file to write discovered WAFs into (JSON format)
    -oc, -output-csv string     file to write discovered WAFs into (CSV format)

  CONFIG:
    -x                      only test one single payload to try to identify the running WAF (useful for in-mass recon)
    -k, -keyword string     keyword to replace in urls (if url doesn't contain keyword no change is done) (default=FUZZ)
    -w, -workers int        number of concurrent workers (default=15)
    -a, -agent string       user agent to include on requests (default=generic agent)
    -p, -proxy string       proxy to send requests through (i.e. http://127.0.0.1:8080)
    -t, -timeout int        milliseconds to wait before timeout (default=4000)
    -c, -color              print colors on output (recommended)
    -q, -quiet              print neither banner nor logging, only print output

  DEBUG:
    -version      show go-recon version
    -h, -help     print help panel
  
Examples:
    gr-waf -u https://example.com -c
    gr-waf -u https://example.com/index.php?foo=FUZZ -p "http://127.0.0.1:8000"
    gr-waf -u https://example.com/index.php?foo=TEST -k TEST -q
    cat urls.txt | gr-waf -t 8000 -c
    `)
}

var csv_info [][]string

// nolint: gocyclo
func main() {
	var url string
	var list string
  var output string
  var json_output string
	var csv_output string
  var one_payload bool
	var keyword string
	var workers int
	var proxy string
	var timeout int
	var user_agent string
	var use_color bool
	var quiet bool
	var version bool
	var help bool
	var stdin bool

	flag.StringVar(&url, "u", "", "")
	flag.StringVar(&url, "url", "", "")
	flag.StringVar(&list, "l", "", "")
	flag.StringVar(&list, "list", "", "")
	flag.StringVar(&output, "o", "", "")
	flag.StringVar(&output, "output", "", "")
	flag.StringVar(&json_output, "oj", "", "")
	flag.StringVar(&json_output, "json-output", "", "")
	flag.StringVar(&csv_output, "oc", "", "")
	flag.StringVar(&csv_output, "output-csv", "", "")
  flag.BoolVar(&one_payload, "x", false, "")
	flag.StringVar(&keyword, "k", "FUZZ", "")
	flag.StringVar(&keyword, "keyword", "FUZZ", "")
	flag.IntVar(&workers, "w", 15, "")
	flag.IntVar(&workers, "workers", 15, "")
	flag.StringVar(&proxy, "proxy", "", "")
	flag.IntVar(&timeout, "t", 4000, "")
	flag.IntVar(&timeout, "timeout", 4000, "")
	flag.StringVar(&user_agent, "a", "Mozilla/5.0 (X11; Linux x86_64; rv:78.0) Gecko/20100101 Firefox/78.0", "")
	flag.StringVar(&user_agent, "agent", "Mozilla/5.0 (X11; Linux x86_64; rv:78.0) Gecko/20100101 Firefox/78.0", "")
	flag.BoolVar(&use_color, "c", false, "")
	flag.BoolVar(&use_color, "color", false, "")
	flag.BoolVar(&quiet, "q", false, "")
	flag.BoolVar(&quiet, "quiet", false, "")
	flag.BoolVar(&version, "version", false, "")
	flag.BoolVar(&help, "h", false, "")
	flag.BoolVar(&help, "help", false, "")
	flag.Parse()

	t1 := core.StartTimer()

	if version {
		fmt.Println("go-recon version:", core.Version())
		os.Exit(0)
	}

	if !quiet {
		fmt.Println(core.Banner())
	}

	if help {
		helpPanel()
		os.Exit(0)
	}

	// Check if stdin has value
	fi, err := os.Stdin.Stat()
	if err != nil {
		log.Fatal(err)
	}

	if fi.Mode()&os.ModeNamedPipe == 0 {
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

	if proxy != "" {
		os.Setenv("HTTP_PROXY", proxy)
		os.Setenv("HTTPS_PROXY", proxy)
	}

	client := core.CreateHttpClient(timeout)

	var json_url string = "https://raw.githubusercontent.com/D3Ext/go-recon/main/utils/waf_vendors.json"
	var m map[string]interface{}

	// send request to waf vendors data (json format)
	req, _ := http.NewRequest("GET", json_url, nil)
  req.Header.Set("User-Agent", user_agent)
	req.Header.Add("Connection", "keep-alive")
	req.Close = true

	resp, err := client.Do(req) // Send request
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Read raw response
	data, _ := ioutil.ReadAll(resp.Body)

	// Parse json
	err = json.Unmarshal(data, &m)
	if err != nil {
		log.Fatal(err)
	}

	if !quiet {
		core.Warning("Use with caution.", use_color)
		if use_color {
			fmt.Println("[" + magenta("*") + "] Parsing WAF vendors data... (" + cyan(strconv.Itoa(len(strings.Split(string(data), "\n")))) + " lines)")
		} else {
			fmt.Println("[*] Parsing WAF vendors data... (" + strconv.Itoa(len(strings.Split(string(data), "\n"))) + " lines)")
		}
	}

  // create TXT output file
	var txt_out *os.File
	if output != "" {
		txt_out, err = os.Create(output)
		if err != nil {
			log.Fatal(err)
		}
	}

  // define variable to hold waf results
  var waf_results []WafResult
  var total_urls []string
  var found_wafs []string

  var payloads_to_use []string
  var payload string

  if one_payload { // only use on payload
    payloads_to_use = []string{"../../../../../etc/passwd"}
  } else { // use multiple payloads
    payloads_to_use = []string{"../../../../../etc/passwd", "' or 1=1-- -", "\"><script>alert(window.origin)</script>"}
  }

	if url != "" {

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
			if value[0] == keyword {
				param_to_test = key
				break
			}
		}

		if !quiet {
			if param_to_test == "" { // enter here if no parameter to test is given
				if use_color {
					fmt.Println("[" + red("!") + "] No parameters detected so analysis may not be accurated enough")
				} else {
					fmt.Println("[!] No parameters detected so analysis may not be accurated enough")
				}
			} else {
				if use_color {
					fmt.Println("["+green("+")+"] Parameter to test detected:", cyan(param_to_test))
				} else {
					fmt.Println("[+] Parameter to test detected:", param_to_test)
				}
			}

			// Check if url is up
			core.Magenta("Testing connection with target url...", use_color)
		}

		req, _ = http.NewRequest("GET", url, nil)
    req.Header.Set("User-Agent", user_agent)
		req.Header.Add("Connection", "keep-alive")
		req.Close = true

		resp, err = client.Do(req) // Send request
		if err != nil {
			log.Fatal(err)
		}

		if !quiet {
			core.Green("Connection succeeded\n", use_color)
      if use_color {
			  core.Magenta("User-Agent: "+cyan(user_agent), use_color)
      } else {
        core.Magenta("User-Agent: "+user_agent, use_color)
      }
		}

    for _, payload := range payloads_to_use {
      if !one_payload {
        fmt.Println()
      }

      var payload_url string // Define url with payload
      if (strings.HasSuffix(url, "/")) && (param_to_test == "") {
        payload_url = url + payload

      } else if (!strings.HasSuffix(url, "/")) && (param_to_test == "") {
        payload_url = url + "/" + payload

      } else if (!strings.HasSuffix(url, "/")) && (param_to_test != "") {
        payload_url = strings.ReplaceAll(url, keyword, payload)
      }

      if !quiet {
        if use_color {
          fmt.Println("["+magenta("*")+"] Payload url:", cyan(payload_url))
        } else {
          fmt.Println("[*] Payload url:", payload_url)
        }
      }

      req, _ = http.NewRequest("GET", payload_url, nil)
      req.Header.Set("User-Agent", user_agent)
      req.Header.Add("Connection", "close")
      req.Close = true

      resp, err = client.Do(req) // Send request
      if err != nil {
        log.Fatal(err)
      }
      defer resp.Body.Close()

      if !quiet {
        if use_color {
          if resp.StatusCode >= 400 {
            fmt.Println("["+red("!")+"] Status code:", cyan(resp.StatusCode))
          } else if resp.StatusCode >= 300 && resp.StatusCode < 400 {
            fmt.Println("["+blue("*")+"] Status code:", cyan(resp.StatusCode))
          } else if resp.StatusCode >= 200 && resp.StatusCode < 300 {
            fmt.Println("["+green("+")+"] Status code:", cyan(resp.StatusCode))
          }
          fmt.Println("[" + magenta("*") + "] Comparing values... (headers, response, cookies, status code)")
        } else {
          if resp.StatusCode >= 400 {
            fmt.Println("[!] Status code:", resp.StatusCode)
          } else if resp.StatusCode >= 300 && resp.StatusCode < 400 {
            fmt.Println("[*] Status code:", resp.StatusCode)
          } else if resp.StatusCode >= 200 && resp.StatusCode < 300 {
            fmt.Println("[+] Status code:", resp.StatusCode)
          }
          fmt.Println("[*] Comparing values... (headers, response, cookies, status code)")
        }
      }

      result, key := checkWaf(m, resp)

      if result == true {
        if !slices.Contains(found_wafs, key) {
          found_wafs = append(found_wafs, key)
        }
      }
    }

    if len(found_wafs) == 0 {
      if use_color {
        fmt.Println("\n[" + red("-") + "] WAF not found")
      } else {
        fmt.Println("\n[-] WAF not found")
      }
    }

    if len(found_wafs) == 1 {
      if use_color {
        fmt.Println("\n["+green("+")+"] WAF found:", cyan(found_wafs[0]))
      } else {
        fmt.Println("\n[+] WAF found:", found_wafs[0])
      }
    } else if len(found_wafs) > 1 {
      fmt.Println()
      core.Green("Multiple WAFs were detected:", use_color)
      for _, w := range found_wafs {
        fmt.Println("  " + w)
      }

      if output != "" {
        _, err = txt_out.WriteString(url + " - " + strings.Join(found_wafs[:], "|") + "\n")
        if err != nil {
          log.Fatal(err)
        }
      }

      total_urls = append(total_urls, url)
      waf_results = append(waf_results, WafResult{Url: url, Waf: strings.Join(found_wafs[:], "|")})
      csv_info = append(csv_info, []string{url, strings.Join(found_wafs[:], "|")})
    }

	} else if (list != "") || (stdin) {

		var f *os.File
		var err error
		if list != "" { // get file descriptor from .txt file or stdin
			f, err = os.Open(list)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

		} else if stdin {
			f = os.Stdin
		}

		urls_c := make(chan string)
		var wg sync.WaitGroup

		for i := 0; i < workers; i++ { // create n workers
			wg.Add(1)

			go func() {
				for u := range urls_c { // get url from channel

          for _, payload := range payloads_to_use {
            if (!strings.HasPrefix(u, "http://")) && (!strings.HasPrefix(u, "https://")) {
              url = "https://" + u
            }

            if strings.HasSuffix(u, "=") {
              url = u + payload

            } else if !strings.HasSuffix(u, "/") {
              url = u + "/" + payload
            }

            req, _ := http.NewRequest("GET", url, nil)
            req.Header.Set("User-Agent", user_agent)          
            req.Header.Add("Connection", "close")
            req.Close = true

            resp, err := client.Do(req) // Send request
            if err != nil {
              log.Fatal(err)
            }
            defer resp.Body.Close()

            // compare web response with WAF vendors data
            result, key := checkWaf(m, resp)

            if result == true {
              found_wafs = append(found_wafs, key)

              if use_color {
                fmt.Println(u, "-", cyan(key))
              } else {
                fmt.Println(u, "-", key)
              }

              if output != "" {
                _, err = txt_out.WriteString(u + " - " + key + "\n")
                if err != nil {
                  log.Fatal(err)
                }
              }

              total_urls = append(total_urls, u)
              waf_results = append(waf_results, WafResult{Url: u, Waf: key})
              csv_info = append(csv_info, []string{u, key})

              break // break loop since the WAF has been discovered
            }
          }

          if len(found_wafs) == 0 {
            if use_color {
              fmt.Println(u, "-", red("Not found"))
            } else {
              fmt.Println(u, "-", "Not found")
            }
          }
				}

				wg.Done() // finish worker
			}()
		}
		fmt.Println()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()

			if line == "" {
				continue
			}

			if (!strings.HasPrefix(line, "http://")) && (!strings.HasPrefix(line, "https://")) {
				line = "https://" + line
			}

			if strings.HasSuffix(line, "=") {
				line = line + payload

			} else if !strings.HasSuffix(line, "/") {
				line = line + "/" + payload
			}

			urls_c <- line
		}

		close(urls_c)
		wg.Wait()
	}

	if json_output != "" {
		json_urls := UrlsInfo{
			Results:  waf_results,
			Length:   len(total_urls),
			Time:     core.TimerDiff(t1).String(),
		}

		json_body, err := json.Marshal(json_urls)
		if err != nil {
			log.Fatal(err)
		}

		json_out, err := os.Create(json_output)
		if err != nil {
			log.Fatal(err)
		}

		_, err = json_out.WriteString(string(json_body))
		if err != nil {
			log.Fatal(err)
		}
	}

	if csv_output != "" {
		csv_out, err := os.Create(csv_output)
		if err != nil {
			log.Fatal(err)
		}

		writer := csv.NewWriter(csv_out)
		defer writer.Flush()

		headers := []string{"url", "waf"}
		writer.Write(headers)

		for _, row := range csv_info {
			writer.Write(row)
		}
	}

	if !quiet {
		fmt.Println()
    if output != "" {
      core.Green("Discovered WAFs written to "+output+" (TXT)", use_color)
    }

    if json_output != "" {
      core.Green("Discovered WAFs written to "+json_output+" (JSON)", use_color)
    }

    if csv_output != "" {
			core.Green("Discovered WAFs written to "+csv_output+" (CSV)", use_color)
		}

		if use_color {
			fmt.Println("["+green("+")+"] Elapsed time:", green(core.TimerDiff(t1)))
		} else {
			fmt.Println("[+] Elapsed time:", core.TimerDiff(t1))
		}
	}
}

func checkWaf(m map[string]interface{}, resp *http.Response) (bool, string) {
  // Define some values which are compared with WAF vendors data to detect them
  var cookies []string
  var headers []string

  code := strconv.Itoa(resp.StatusCode)   // Get status code
  page, _ := ioutil.ReadAll(resp.Body)    // Parse page content
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

  // define variable to determine whether the WAF has been discovered or not
  var result float32 = 0

  for key, value := range m { // iterate over json data

    code_to_check, _ := find(value, "code") // check response status code
    if code_to_check.(string) != "" {
      res, err := regexp.MatchString(code_to_check.(string), code)
      if err != nil {
        continue
      }

      if res {
        result += 0.5
      }
    }

    page_to_check, _ := find(value, "page") // check specific strings
    if page_to_check.(string) != "" {
      res, err := regexp.MatchString(page_to_check.(string), string(page))
      if err != nil {
        continue
      }
      //fmt.Println(res)
      //fmt.Println(page_to_check.(string))
      //fmt.Println(string(page))

      if res {
        result += 1
      }
    }

    headers_to_check, _ := find(value, "headers") // check response headers
    if headers_to_check.(string) != "" {
      for _, h := range headers {
        res, err := regexp.MatchString(headers_to_check.(string), h)
        if err != nil {
          continue
        }

        if res {
          result += 1
        }
      }
    }

    cookies_to_check, _ := find(value, "cookie") // check present cookies 
    if cookies_to_check.(string) != "" {
      for _, c := range cookies {
        res, err := regexp.MatchString(cookies_to_check.(string), c)
        if err != nil {
          continue
        }

        if res {
          result += 1
        }
      }
    }

    // check if WAF was found to skip remaining checks
    if result >= 1 {
      return true, key
    }

    result = 0
  }

  return false, ""
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


