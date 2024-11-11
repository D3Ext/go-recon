package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
	wappalyzer "github.com/projectdiscovery/wappalyzergo"
	"io/ioutil"
	"log"
	"net/http"
	nu "net/url"
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

type UrlsInfo struct {
  Urls   []string `json:"urls"`
	Length int      `json:"length"`
	Time   string   `json:"time"`
}

func helpPanel() {
	fmt.Println(`Usage of gr-probe:
  INPUT:
    -d, -domain string    domain/url/host to probe for working http and https servers
    -l, -list string      file containing a list of domains/urls to probe (one url per line)

  OUTPUT:
    -o, -output string          file to write active urls into (TXT format)
    -oj, -output-json string    file to write active urls into (JSON format)
    -oc, -output-csv string     file to write active urls into (CSV format)

  FILTER:
    -fs, -filter-string string    filter urls that match an especific string on response body
    -fc, -filter-code int[]       filter urls that match an especific status codes (separated by comma) (i.e. 403,401)

  INFO:
    -status       show status code of each request
    -title        show title of each request
    -location     show redirect location if exists
    -body         show response body instead of active host

  CONFIG:
    -x                    check active domain/host but don't include http:// or https://
    -techs                use go-wappalyzer to detect technologies running on each url
    -nr, -no-redirects    don't follow http redirects
    -s, -skip             don't check http if https is working (default=disabled)
    -m, -method string    requests method (i.e. POST)
    -w, -workers int      number of concurrent workers (split same amount between http and https) (default=20)
    -a, -agent string     user agent to include on requests (default=generic agent)
    -p, -proxy string     proxy to send requests through (i.e. http://127.0.0.1:8080)
    -t, -timeout int      milliseconds to wait before each request timeout (default=4000)
    -c, -color            use color on output
    -q, -quiet            print neither banner nor logging, only print output

  DEBUG:
    -version      show go-recon version
    -h, -help     print help panel
  
Examples:
    gr-probe -l domains.txt -w 15 -o results.txt
    gr-probe -l domains.txt -status -title -location -c
    gr-probe -l urls.txt -techs -c
    cat domains.txt | gr-probe -skip -c
    cat domains.txt | gr-probe -quiet -o urls.txt
    `)
}

func printWithColor(n int, use_color bool) string {
	if use_color {
		if n >= 200 && n < 300 {
			return green(strconv.Itoa(n))
		} else if n >= 300 && n < 400 {
			return cyan(strconv.Itoa(n))
		} else if n >= 400 && n < 500 {
			return red(strconv.Itoa(n))
		} else if n >= 500 && n < 600 {
			return yellow(strconv.Itoa(n))
		} else {
			return strconv.Itoa(n)
		}
	} else {
		return strconv.Itoa(n)
	}
}

var csv_info [][]string

// nolint: gocyclo
func main() {
	var domain string
	var list string
	var skip_http bool
	var method string
	var workers int
	var proxy string
	var timeout int
	var user_agent string
	var output string
	var json_output string
	var csv_output string
	var use_color bool
	var string_to_search string
	var status_code_to_search int
  var dont_include_method bool
	var techs bool
	var no_redirects bool
	var show_body bool
	var status_code bool
	var title bool
	var location bool
	var quiet bool
	var version bool
	var help bool
	var stdin bool

	flag.StringVar(&domain, "d", "", "")
	flag.StringVar(&domain, "domain", "", "")
	flag.StringVar(&list, "l", "", "")
	flag.StringVar(&list, "list", "", "")
	flag.BoolVar(&skip_http, "s", false, "")
	flag.BoolVar(&skip_http, "skip", false, "")
	flag.StringVar(&method, "m", "GET", "")
	flag.StringVar(&method, "method", "GET", "")
	flag.IntVar(&workers, "w", 20, "")
	flag.IntVar(&workers, "workers", 20, "")
	flag.IntVar(&timeout, "t", 4000, "")
	flag.IntVar(&timeout, "timeout", 4000, "")
	flag.StringVar(&output, "o", "", "")
	flag.StringVar(&output, "output", "", "")
	flag.StringVar(&json_output, "oj", "", "")
	flag.StringVar(&json_output, "output-json", "", "")
	flag.StringVar(&csv_output, "oc", "", "")
	flag.StringVar(&csv_output, "output-csv", "", "")
	flag.StringVar(&proxy, "proxy", "", "")
	flag.StringVar(&user_agent, "a", "Mozilla/5.0 (X11; Linux x86_64; rv:78.0) Gecko/20100101 Firefox/78.0", "")
	flag.StringVar(&user_agent, "agent", "Mozilla/5.0 (X11; Linux x86_64; rv:78.0) Gecko/20100101 Firefox/78.0", "")
	flag.StringVar(&string_to_search, "fs", "", "")
	flag.StringVar(&string_to_search, "filter-string", "", "")
	flag.IntVar(&status_code_to_search, "fc", 0, "")
	flag.IntVar(&status_code_to_search, "filter-code", 0, "")
	flag.BoolVar(&no_redirects, "nr", false, "")
	flag.BoolVar(&no_redirects, "no-redirects", false, "")
  flag.BoolVar(&dont_include_method, "x", false, "")
	flag.BoolVar(&techs, "techs", false, "")
	flag.BoolVar(&show_body, "body", false, "")
	flag.BoolVar(&status_code, "status", false, "")
	flag.BoolVar(&title, "title", false, "")
	flag.BoolVar(&location, "location", false, "")
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
	if (domain == "") && (list == "") && (!stdin) {
		helpPanel()
		os.Exit(0)
	}

	if (domain != "") && (list != "") {
		helpPanel()
		core.Red("You can't use (-d) and (-l) at same time", use_color)
		os.Exit(0)
	}

	if proxy != "" {
		os.Setenv("HTTP_PROXY", proxy)
		os.Setenv("HTTPS_PROXY", proxy)
	}

	var status_codes []int
	for _, sc := range strings.Split(strings.TrimSpace(strconv.Itoa(status_code_to_search)), ",") {
		s, err := strconv.Atoi(sc)
		if err != nil {
			log.Fatal(err)
		}

		status_codes = append(status_codes, s)
	}

	var txt_out *os.File
	if output != "" {
		txt_out, err = os.Create(output)
		if err != nil {
			log.Fatal(err)
		}
	}

	var client *http.Client
	if no_redirects {
		client = core.CreateHttpClient(timeout) // Create http client via auxiliary function (don't follow redirects)
	} else {
		client = core.CreateHttpClientFollowRedirects(timeout) // Create http client via auxiliary function (follow redirects)
	}

	// Create channels
	https_c := make(chan string)
	http_c := make(chan string)
	out := make(chan string)
	var counter int
	var probed_urls []string

	if !quiet {
		core.Warning("Use with caution.", use_color)
		core.Magenta("Concurrent workers: "+strconv.Itoa(workers), use_color)
		if proxy != "" {
			core.Magenta("Proxy: "+proxy, use_color)
		}
		core.Magenta("Sending requests to given servers...\n", use_color)
	}

	if (domain != "") || (list != "") || (stdin) { // Enter here if list was given or if stdin has value

		var f *os.File
		var err error

		if list != "" {
			f, err = os.Open(list)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

		} else if stdin {
			f = os.Stdin
		}

		// start with HTTPS
		var https_wg sync.WaitGroup      // create wait group for https urls
		for i := 0; i < workers/2; i++ { // split workers and launch half of them
			https_wg.Add(1)

			go func() {
				for url := range https_c { // receive urls from channel
					req, err := http.NewRequest(method, "https://"+url, nil)
					if err != nil { // handle error and continue
						continue
					}

          req.Header.Set("User-Agent", user_agent)
					req.Header.Add("Connection", "close")
					req.Close = true

					resp, err := client.Do(req)
					if err == nil {
						doc, err := goquery.NewDocumentFromReader(resp.Body)
						if err != nil {
							log.Fatal(err)
						}
						defer resp.Body.Close()

						body := doc.Selection.Text()
						title_str := doc.Find("title").Text()

						var format_str string

						if (status_code_to_search == 0 || status_code_to_search == resp.StatusCode) && !show_body {
							if string_to_search != "" {
								if !strings.Contains(string(body), string_to_search) {
									continue
								}
							}

							resp_l := resp.Header.Get("Location")

							format_str = formatInfo("https://"+url, use_color, status_code, resp.StatusCode, title, title_str, location, resp_l)

              if (dont_include_method) {
  							probed_urls = append(probed_urls, url)
  							csv_info = append(csv_info, []string{url, strconv.Itoa(resp.StatusCode), title_str})
              } else {
  							probed_urls = append(probed_urls, "https://"+url)
  							csv_info = append(csv_info, []string{"https://"+url, strconv.Itoa(resp.StatusCode), title_str})
              }

							if techs {
								format_str = format_str + getTechs(resp, use_color)
							}

							out <- format_str
							counter += 1

						} else if (status_code_to_search == 0 || status_code_to_search == resp.StatusCode) && show_body {
							if string_to_search != "" {
								if !strings.Contains(string(body), string_to_search) {
									continue
								}
							}

							out <- string(body)
							counter += 1
						}

						if skip_http { // skip to next url so that it is not sent through http channel
							continue
						}
					}

					http_c <- url
				}

				https_wg.Done()
			}()
		}

		// now continue with HTTP
		var http_wg sync.WaitGroup       // create http wait group
		for i := 0; i < workers/2; i++ { // launch rest of workers
			http_wg.Add(1)

			go func() {
				for url := range http_c { // receive urls from channel
					req, err := http.NewRequest(method, "http://"+url, nil)
					if err != nil {
						continue
					}

          req.Header.Set("User-Agent", user_agent)
					req.Header.Add("Connection", "close")
					req.Close = true

					resp, err := client.Do(req)
					if err == nil {
						doc, err := goquery.NewDocumentFromReader(resp.Body)
						if err != nil {
							log.Fatal(err)
						}
						defer resp.Body.Close()

						body := doc.Selection.Text()
						title_str := doc.Find("title").Text()

						var format_str string

						if (status_code_to_search == 0 || status_code_to_search == resp.StatusCode) && !show_body {
							if string_to_search != "" {
								if !strings.Contains(string(body), string_to_search) {
									continue
								}
							}

							resp_l := resp.Header.Get("Location")

							format_str = formatInfo("http://"+url, use_color, status_code, resp.StatusCode, title, title_str, location, resp_l)

              if (dont_include_method) {
  							probed_urls = append(probed_urls, url)
  							csv_info = append(csv_info, []string{url, strconv.Itoa(resp.StatusCode), title_str})
              } else {
  							probed_urls = append(probed_urls, "http://"+url)
  							csv_info = append(csv_info, []string{"http://"+url, strconv.Itoa(resp.StatusCode), title_str})
              }

							if techs {
								format_str = format_str + getTechs(resp, use_color)
							}

							out <- format_str
							counter += 1

						} else if (status_code_to_search == 0 || status_code_to_search == resp.StatusCode) && show_body {
							if string_to_search != "" {
								if !strings.Contains(string(body), string_to_search) {
									continue
								}
							}

							out <- string(body)
							counter += 1
						}
					}
				}

				http_wg.Done()
			}()
		}

		// close http channel when https workers finish
		go func() {
			https_wg.Wait()
			close(http_c)
		}()

		var out_wg sync.WaitGroup // create output wait group
		out_wg.Add(1)             // 1 worker

		go func() {
			for o := range out { // receive urls to print from output channel
        if (dont_include_method) { // check if (-x) parameter was given to remove http:// and https:// from results
          o = strings.TrimPrefix(o, "https://")
          o = strings.TrimPrefix(o, "http://")
        }
				fmt.Println(o)
        url := strings.Split(o, " ")[0]

				if output != "" { // if output file is given, write urls into it
					_, err = txt_out.WriteString(url + "\n")
					if err != nil {
						log.Fatal(err)
					}
				}
			}

			out_wg.Done()
		}()

		go func() { // close output channel when http work group finish
			http_wg.Wait()
			close(out)
		}()

		// load urls to HTTPS channel
		if (list != "") || (stdin) {
			scanner := bufio.NewScanner(f)
			for scanner.Scan() { // iterate over every single line
				line := scanner.Text()

				if (strings.HasPrefix(line, "http://")) || (strings.HasPrefix(line, "https://")) { // Check if domain is valid
					u, _ := nu.Parse(line)
					https_c <- strings.TrimPrefix(line, u.Scheme+"://") // remove scheme

				} else if strings.Contains(line, ".") {
					https_c <- line
				}
			}
		} else {
			if (strings.HasPrefix(domain, "http://")) || (strings.HasPrefix(domain, "https://")) { // Check if domain is valid
				u, _ := nu.Parse(domain)
				https_c <- strings.TrimPrefix(domain, u.Scheme+"://") // remove scheme

			} else if strings.Contains(domain, ".") {
				https_c <- domain
			}
		}

		close(https_c)
		out_wg.Wait() // wait for output
	}

	if json_output != "" {
		json_urls := UrlsInfo{
			Urls:   probed_urls,
			Length: len(probed_urls),
			Time:   core.TimerDiff(t1).String(),
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

		headers := []string{"url", "status", "title"}

		writer.Write(headers)
		for _, row := range csv_info {
			writer.Write(row)
		}
	}

	// finally add some logging to aid users
	if !quiet {
		if counter >= 1 {
			fmt.Println()
			if output != "" {
				core.Green("Urls written to "+output+" (TXT)", use_color)
			}

			if json_output != "" {
				core.Green("Urls written to "+json_output+" (JSON)", use_color)
			}

			if csv_output != "" {
				core.Green("Urls written to "+csv_output+" (CSV)", use_color)
			}

			if use_color {
				fmt.Println("["+green("+")+"]", cyan(counter), "urls are active")
			} else {
				fmt.Println("[+]", counter, "urls are active")
			}

		} else {
			if string_to_search == "" && status_code_to_search == 0 {
				core.Red("No url is active", use_color)
			} else {
				core.Red("No url found", use_color)
			}
		}

		if use_color {
			fmt.Println("["+green("+")+"] Elapsed time:", green(core.TimerDiff(t1)))
		} else {
			fmt.Println("[+] Elapsed time:", core.TimerDiff(t1))
		}
	}
}

func getTechs(resp *http.Response, use_color bool) string {
	var final_str string

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	wappalyzer_client, err := wappalyzer.New()
	if err != nil {
		log.Fatal(err)
	}

	var techs_array []string
	techs := wappalyzer_client.Fingerprint(resp.Header, data)

	if len(techs) >= 1 {
		final_str = " - ["
		for k := range techs {
			techs_array = append(techs_array, k)
		}

		for _, v := range techs_array {
			if techs_array[len(techs_array)-1] == v {
				if use_color {
					final_str = fmt.Sprintf("%s%s]", final_str, cyan(v))
				} else {
					final_str = fmt.Sprintf("%s%s]", final_str, v)
				}
			} else {
				if use_color {
					final_str = fmt.Sprintf("%s%s, ", final_str, cyan(v))
				} else {
					final_str = fmt.Sprintf("%s%s, ", final_str, v)
				}
			}
		}
	}

	return final_str
}

func formatInfo(url string, use_color bool, status_code bool, resp_code int, title bool, title_str string, location bool, resp_l string) string {
	if status_code {
		url = url + " - [" + printWithColor(resp_code, use_color) + "]"
	}

	if (title) && (title_str != "") {
		url = url + " - [Title: " + title_str + "]"
	}

	if (location) && (resp_l != "") {
		url = url + " - [Location: " + resp_l + "]"
	}

	return url
}

func checkStatusCodes(codes_to_filter []int, code int) bool {
	for _, c := range codes_to_filter {
		if c == code {
			return true
		}
	}

	return false
}
