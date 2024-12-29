package main

import (
	"bufio"
  "encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/fatih/color"
  nurl "net/url"
	//tld "github.com/jpillora/go-tld"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/D3Ext/go-recon/core"
)

var red func(a ...interface{}) string = color.New(color.FgRed).SprintFunc()
var cyan func(a ...interface{}) string = color.New(color.FgCyan).SprintFunc()
var green func(a ...interface{}) string = color.New(color.FgGreen).SprintFunc()
var magenta func(a ...interface{}) string = color.New(color.FgMagenta).SprintFunc()
var yellow func(a ...interface{}) string = color.New(color.FgYellow).SprintFunc()

type RedirectsInfo struct {
	Urls   []string `json:"urls"`
	Length int      `json:"length"`
	Time   string   `json:"time"`
}

func helpPanel() {
	fmt.Println(`Usage of gr-openredirects:
  INPUT:
    -u, -url string       url to check open redirect (i.e. https://example.com/?foo=FUZZ)
    -l, -list string      file containing a list of urls to check open redirects (one url per line)

  OUTPUT:
    -o, -output string          file to write vulnerable urls into (TXT format)
    -oj, -output-json string    file to write vulnerable urls into (JSON format)
    -oc, -output-csv string     file to write vulnerable urls into (CSV format)

  PAYLOADS:
    -k, -keyword string           keyword to replace in urls with payloads (default=FUZZ)
    -pl, -payloads-list string    file containing a list of payloads (if not especified, default payloads are used)
    -s, -skip                     only test most common payloads (useful for in-mass testing)

  CONFIG:
    -w, -workers int        number of concurrent workers (default=15)
    -m, -method string      requests method (GET, POST, PUT...)
    -H, -header string      include custom headers (separated by semicolon) on HTTP requests
    -a, -agent string       user agent to include on requests (default=generic agent)
    -p, -proxy string       proxy to send requests through (i.e. http://127.0.0.1:8080)
    -t, -timeout int        milliseconds to wait before each request timeout (default=5000)
    -c, -color              use color on output
    -q, -quiet              print neither banner nor logging, only print output

  DEBUG:
    -version      show go-recon version
    -h, -help     print help panel

Examples:
    gr-openredirects -u https://example.com/?foo=FUZZ -c
    gr-openredirects -u https://example.com/?foo=TEST -k TEST
    gr-openredirects -l urls.txt -skip
    gr-openredirects -l urls.txt -pl payloads.txt -o vulnerable.txt
    cat urls.txt | gr-openredirects
    `)
}

var found_redirects []string

// nolint: gocyclo
func main() {
	var url string
	var list string
	var keyword string
	var payloads_list string
	var workers int
	var skip bool
	var method string
  var header string
	var proxy string
	var timeout int
	var user_agent string
	var output string
	var json_output string
  var csv_output string
	var use_color bool
	var quiet bool
	var version bool
	var help bool
	var stdin bool

	flag.StringVar(&url, "u", "", "")
	flag.StringVar(&url, "url", "", "")
	flag.StringVar(&list, "l", "", "")
	flag.StringVar(&list, "list", "", "")
	flag.StringVar(&keyword, "k", "FUZZ", "")
	flag.StringVar(&keyword, "keyword", "FUZZ", "")
	flag.StringVar(&payloads_list, "pl", "", "")
	flag.StringVar(&payloads_list, "payloads-list", "", "")
	flag.BoolVar(&skip, "s", false, "")
	flag.BoolVar(&skip, "skip", false, "")
	flag.IntVar(&workers, "w", 10, "")
	flag.IntVar(&workers, "workers", 10, "")
	flag.StringVar(&method, "m", "GET", "")
	flag.StringVar(&method, "method", "GET", "")
  flag.StringVar(&header, "H", "", "")
  flag.StringVar(&header, "header", "", "")
	flag.StringVar(&user_agent, "a", "Mozilla/5.0 (X11; Linux x86_64; rv:78.0) Gecko/20100101 Firefox/78.0", "")
	flag.StringVar(&user_agent, "agent", "Mozilla/5.0 (X11; Linux x86_64; rv:78.0) Gecko/20100101 Firefox/78.0", "")
	flag.StringVar(&proxy, "p", "", "")
	flag.StringVar(&proxy, "proxy", "", "")
	flag.IntVar(&timeout, "t", 4000, "")
	flag.IntVar(&timeout, "timeout", 4000, "")
	flag.StringVar(&output, "o", "", "")
	flag.StringVar(&output, "output", "", "")
	flag.StringVar(&json_output, "oj", "", "")
	flag.StringVar(&json_output, "output-json", "", "")
  flag.StringVar(&csv_output, "oc", "", "")
  flag.StringVar(&csv_output, "output-csv", "", "")
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
		fmt.Println()
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
		helpPanel()
		os.Exit(0)
	}

	if payloads_list != "" && skip {
		helpPanel()
		core.Red("You can't use (-pl) and (-skip) at the same time", use_color)
		os.Exit(0)
	}

	if proxy != "" {
		os.Setenv("HTTP_PROXY", proxy)
		os.Setenv("HTTPS_PROXY", proxy)
	}

	var txt_out *os.File
	if output != "" {
		txt_out, err = os.Create(output)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Create requests client
	client := core.CreateHttpClientFollowRedirects(timeout)

  var csv_info [][]string
	var payloads []string

	if skip { // set unique payload if user only want to use one
		payloads = []string{"https://bing.com", "bing.com", "//bing.com"}

	} else if payloads_list == "" { // set default payloads
		payloads = core.GetPayloads()

	} else { // get payloads from custom list
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
      parse, err := nurl.Parse(p)
      if err != nil {
        continue
      }

      redirectTarget = parse.Host

			break
		}
	}

	if !quiet {
		core.Warning("Use with caution.", use_color)
		core.Magenta("Concurrent workers: "+strconv.Itoa(workers), use_color)
		if proxy != "" {
			core.Magenta("Proxy: "+proxy, use_color)
		}

		if skip {
			core.Magenta("Payload to use: https://bing.com", use_color)
		}

		core.Magenta("Checking open redirects with "+strconv.Itoa(len(payloads))+" payloads\n", use_color)
	}

	var counter int
	urls_c := make(chan string) // Create channel which holds urls to use
	payloads_c := make(chan string)
	var wg sync.WaitGroup

	if url != "" {
		if !strings.Contains(url, keyword) {
			core.Red("Url doesn't contain keyword ("+keyword+")", use_color)
			os.Exit(0)
		}

		_, err = http.Get(url)
		if err != nil {
			log.Fatal(err)
		}

		for i := 0; i < workers; i++ {
			wg.Add(1)

			go func() {
				for payload := range payloads_c { // test all payloads on current url
					new_url := strings.Replace(url, keyword, payload, -1) // Replace keyword in url with each payload
					req, err := http.NewRequest(method, new_url, nil)
					if err != nil {
						continue
					}

          if header != "" {
            for _, h := range strings.Split(header, ";") {
              header_name := strings.Split(h, ":")[0]
              header_value := strings.ReplaceAll(strings.Split(h, ":")[1], " ", "")
              req.Header.Set(header_name, header_value)
            }
          }

          req.Header.Set("User-Agent", user_agent)
					req.Header.Add("Connection", "close")
					req.Close = true

					resp, err := client.Do(req) // Send requests with out custom client config
					if err != nil {
						continue
					}
					defer resp.Body.Close()

					if resp.StatusCode == http.StatusOK {
            finalURL, err := nurl.Parse(resp.Request.URL.String())
            if err != nil {
              continue
            }

            if finalURL.Host == redirectTarget || finalURL.Hostname() == redirectTarget {
							fmt.Println(new_url)
							found_redirects = append(found_redirects, new_url)
              csv_info = append(csv_info, []string{url, payload})
							counter += 1

							if output != "" { // Write url with payload to output file
								_, err = txt_out.WriteString(new_url + "\n")
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

			go func() {
				for u := range urls_c { // iterate over urls
					current_url := u

					for _, payload := range payloads { // test all payloads current url
						new_url := strings.Replace(current_url, keyword, payload, -1) // Replace keyword in url with each payload
						req, err := http.NewRequest(method, new_url, nil)
						if err != nil {
							continue
						}

            if header != "" {
              for _, h := range strings.Split(header, ";") {
                header_name := strings.Split(h, ":")[0]
                header_value := strings.ReplaceAll(strings.Split(h, ":")[1], " ", "")
                req.Header.Set(header_name, header_value)
              }
            }

            req.Header.Set("User-Agent", user_agent)
						req.Header.Add("Connection", "close")
						req.Close = true

						resp, err := client.Do(req) // Send requests with out custom client config
						if err != nil {
							continue
						}
						defer resp.Body.Close()

						if resp.StatusCode == http.StatusOK {
              finalURL, err := nurl.Parse(resp.Request.URL.String())
              if err != nil {
                continue
              }

              if finalURL.Host == redirectTarget || finalURL.Hostname() == redirectTarget {
								fmt.Println(new_url)
								found_redirects = append(found_redirects, new_url)
                csv_info = append(csv_info, []string{url, payload})
								counter += 1

								if output != "" { // Write url with payload to output file
									_, err = txt_out.WriteString(new_url + "\n")
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
		if list != "" {
			f, err = os.Open(list)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

		} else if stdin {
			f = os.Stdin
		}

		scanner := bufio.NewScanner(f)
		for scanner.Scan() { // Iterate over every single line
			line := scanner.Text()

			if line != "" {
				urls_c <- line
			}
		}

		close(urls_c) // Close channel
		wg.Wait()     // and wait for urls workers wait group
	}

	if json_output != "" {
		json_redirects := RedirectsInfo{
			Urls:   found_redirects,
			Length: len(found_redirects),
			Time:   core.TimerDiff(t1).String(),
		}

		json_body, err := json.Marshal(json_redirects)
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

		headers := []string{"url", "payload"}

		writer.Write(headers)
		for _, row := range csv_info {
			writer.Write(row)
		}
	}

	// Finally some logging to aid users
	if !quiet {
		if counter >= 1 {
			fmt.Println()
			core.Green(strconv.Itoa(counter)+" open redirects found", use_color)
		} else {
			core.Red("No open redirect found!", use_color)
		}

		if counter >= 1 { // Check if at least one url was vulnerable to open redirect
			if output != "" {
				core.Green("Urls written to "+output+" (TXT)", use_color)
			}

			if json_output != "" {
				core.Green("Urls written to "+json_output+" (JSON)", use_color)
			}

      if csv_output != "" {
        core.Green("Urls written to "+csv_output+" (CSV)", use_color)
      }
		}

		if use_color {
			if output != "" || counter >= 1 {
				fmt.Println("["+green("+")+"] Elapsed time:", green(core.TimerDiff(t1)))
			} else {
				fmt.Println("\n["+green("+")+"] Elapsed time:", green(core.TimerDiff(t1)))
			}
		} else {
			if output != "" || counter >= 1 {
				fmt.Println("[+] Elapsed time:", core.TimerDiff(t1))
			} else {
				fmt.Println("\n[+] Elapsed time:", core.TimerDiff(t1))
			}
		}
	}
}
