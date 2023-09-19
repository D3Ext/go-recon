package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
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
    -d)         domain to probe for working http and https servers
    -l)         file containing a list of domains to probe (one per line)
    -x)         don't check http if https is working (default=disabled)
    -m)         requests method (GET, POST, PUT...)
    -w)         number of concurrent workers (split same amount between http and https) (default=20)
    -a)         user agent to include on requests (default=none)
    -t)         milliseconds to wait before each request timeout (default=4000)
    -sc)        filter for an especific status code (i.e. 403)
    --status)   show status code of each request
    --title)    show title of each request
    --location) show redirect location if exists
    -o)         file to write active urls into
    -oj)        file to write active urls into (JSON format)
    -c)         use color on output
    -q)         don't print banner nor logging, only output
    -h)         print help panel
  
Examples:
    gr-probe -l domains.txt -w 15 -o results.txt
    gr-probe -l domains.txt --status --title --location -c
    cat domains.txt | gr-probe -x -c
    cat domains.txt | gr-probe -q -o urls.txt
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

func main() {
	var domain string
	var list string
	var prefer_https bool
	var method string
	var workers int
	var timeout int
	var user_agent string
	var output string
	var json_output string
	var use_color bool
	var status_code_to_search int
	var status_code bool
	var title bool
	var location bool
	var quiet bool
	var help bool
	var stdin bool

	flag.StringVar(&domain, "d", "", "domain to probe for working http and https servers")
	flag.StringVar(&list, "l", "", "file containing a list of domain to probe (one per line)")
	flag.BoolVar(&prefer_https, "x", false, "don't check http if https is working")
	flag.StringVar(&method, "m", "GET", "requests method (GET, POST, PUT...)")
	flag.IntVar(&workers, "w", 20, "number of concurrent workers (split same amount between http and https)")
	flag.IntVar(&timeout, "t", 4000, "milliseconds to wait before each request timeout (default=4000)")
	flag.StringVar(&output, "o", "", "file to write active urls into")
	flag.StringVar(&json_output, "oj", "", "file to write active urls into (JSON format)")
	flag.StringVar(&user_agent, "a", "", "user agent to include on requests (default=none)")
	flag.IntVar(&status_code_to_search, "sc", 0, "filter for an especific status code (i.e. 403)")
	flag.BoolVar(&status_code, "status", false, "show status code of each request")
	flag.BoolVar(&title, "title", false, "show title of each request")
	flag.BoolVar(&location, "location", false, "show redirect location")
	flag.BoolVar(&use_color, "c", false, "use color on output")
	flag.BoolVar(&quiet, "q", false, "don't print banner nor logging, only output")
	flag.BoolVar(&help, "h", false, "print help panel")
	flag.Parse()

	t1 := core.StartTimer()

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

	var out_f *os.File
	if output != "" {
		out_f, err = os.Create(output)
		if err != nil {
			log.Fatal(err)
		}
	} else if json_output != "" {
		out_f, err = os.Create(json_output)
		if err != nil {
			log.Fatal(err)
		}
	}

	client := core.CreateHttpClient(timeout) // Create http client via auxiliary function

	// Create channels
	https_c := make(chan string)
	http_c := make(chan string)
	out := make(chan string)
	var counter int
	var probed_urls []string

	if domain != "" { // Enter here if domain was given
		// send request to domain HTTPS server
		req, err := http.NewRequest(method, "https://"+domain, nil)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Add("Connection", "close")
		if user_agent != "" {
			req.Header.Set("User-Agent", user_agent)
		}
		req.Close = true

		resp, err := client.Do(req) // send request
		if err == nil {
			doc, err := goquery.NewDocumentFromReader(resp.Body)
			if err != nil {
				log.Fatal(err)
			}

			title_str := doc.Find("title").First().Text()

			if status_code_to_search == 0 { // don't look for an especific status code
				fmt.Print("https://" + domain)
				if status_code {
					fmt.Print(" - [" + printWithColor(resp.StatusCode, use_color) + "]")
				}

				if (title) && (title_str != "") {
					fmt.Print(" - [Title: " + title_str + "]")
				}

				if (location) && (resp.Header.Get("Location") != "") {
					fmt.Print(" - [Location: " + resp.Header.Get("Location") + "]")
				}

				fmt.Println()
				probed_urls = append(probed_urls, "https://"+domain)

			} else if status_code_to_search == resp.StatusCode {
				fmt.Println("https://" + domain)
				probed_urls = append(probed_urls, "https://"+domain)
			}

			defer resp.Body.Close()
			counter += 1
		}

		// now send request to domain HTTP server
		req, err = http.NewRequest(method, "http://"+domain, nil)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Add("Connection", "close")
		if user_agent != "" {
			req.Header.Set("User-Agent", user_agent)
		}
		req.Close = true

		resp, err = client.Do(req) // send request
		if err == nil {
			doc, err := goquery.NewDocumentFromReader(resp.Body)
			if err != nil {
				log.Fatal(err)
			}

			title_str := doc.Find("title").Text()

			if status_code_to_search == 0 { // don't look for an especific status code
				fmt.Print("http://" + domain)
				if status_code {
					fmt.Print(" - [" + printWithColor(resp.StatusCode, use_color) + "]")
				}

				if (title) && (title_str != "") {
					fmt.Print(" - [Title: " + title_str + "]")
				}

				if (location) && (resp.Header.Get("Location") != "") {
					fmt.Print(" - [Location: " + resp.Header.Get("Location") + "]")
				}

				fmt.Println()
				probed_urls = append(probed_urls, "http://"+domain)

			} else if status_code_to_search == resp.StatusCode {
				fmt.Println("http://" + domain)
				probed_urls = append(probed_urls, "http://"+domain)
			}

			defer resp.Body.Close()
			counter += 1
		}

	} else if (list != "") || (stdin) { // Enter here if list was given or if stdin has value

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

					req.Header.Add("Connection", "close")
					if user_agent != "" {
						req.Header.Set("User-Agent", user_agent)
					}
					req.Close = true

					resp, err := client.Do(req)
					if err == nil {
						doc, err := goquery.NewDocumentFromReader(resp.Body)
						if err != nil {
							log.Fatal(err)
						}

						title_str := doc.Find("title").Text()

						if status_code_to_search == 0 {
							format_str := "https://" + url
							probed_urls = append(probed_urls, "https://"+url)

							if status_code {
								format_str = format_str + " - [" + printWithColor(resp.StatusCode, use_color) + "]"
							}

							if (title) && (title_str != "") {
								format_str = format_str + " - [Title: " + title_str + "]"
							}

							if (location) && (resp.Header.Get("Location") != "") {
								format_str = format_str + " - [Location: " + resp.Header.Get("Location") + "]"
							}

							out <- format_str

						} else if status_code_to_search == resp.StatusCode {
							out <- "https://" + url
							probed_urls = append(probed_urls, "https://"+url)
						}

						defer resp.Body.Close()
						counter += 1

						if prefer_https { // skip to next url so it isn't sent through http channel
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

					req.Header.Add("Connection", "close")
					if user_agent != "" {
						req.Header.Set("User-Agent", user_agent)
					}
					req.Close = true

					resp, err := client.Do(req)
					if err == nil {
						doc, err := goquery.NewDocumentFromReader(resp.Body)
						if err != nil {
							log.Fatal(err)
						}

						title_str := doc.Find("title").Text()

						if status_code_to_search == 0 {
							format_str := "http://" + url
							probed_urls = append(probed_urls, "http://"+url)

							if status_code {
								format_str = format_str + " - [" + printWithColor(resp.StatusCode, use_color) + "]"
							}

							if (title) && (title_str != "") {
								format_str = format_str + " - [Title: " + title_str + "]"
							}

							if (location) && (resp.Header.Get("Location") != "") {
								format_str = format_str + " - [Location: " + resp.Header.Get("Location") + "]"
							}

							out <- format_str

						} else if status_code_to_search == resp.StatusCode {
							out <- "http://" + url
							probed_urls = append(probed_urls, "http://"+url)
						}

						defer resp.Body.Close()
						counter += 1
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
				url := strings.Split(o, " ")[0]
				fmt.Println(o)

				if output != "" { // if output file is given, write urls into it
					_, err = out_f.WriteString(url + "\n")
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

		_, err = out_f.WriteString(string(json_body))
		if err != nil {
			log.Fatal(err)
		}
	}

	// finally add some logging to aid users
	if !quiet {
		if counter >= 1 {
			fmt.Println()
			if output != "" {
				core.Green("Urls written to "+output, use_color)
			} else if json_output != "" {
				core.Green("Urls written to "+json_output, use_color)
			}

			if use_color {
				fmt.Println("["+green("+")+"]", cyan(counter), "urls are active")
			} else {
				fmt.Println("[+]", counter, "urls are active")
			}

		} else {
			core.Red("No url is active", use_color)
		}

		if use_color {
			fmt.Println("["+green("+")+"] Elapsed time:", green(core.TimerDiff(t1)))
		} else {
			fmt.Println("[+] Elapsed time:", core.TimerDiff(t1))
		}
	}
}
