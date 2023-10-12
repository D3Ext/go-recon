package main

import (
	"bufio"
	"crypto/tls"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/gocolly/colly/v2"
	"log"
	"net/http"
	nurl "net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/D3Ext/go-recon/core"
)

var red func(a ...interface{}) string = color.New(color.FgRed).SprintFunc()
var cyan func(a ...interface{}) string = color.New(color.FgCyan).SprintFunc()
var green func(a ...interface{}) string = color.New(color.FgGreen).SprintFunc()
var magenta func(a ...interface{}) string = color.New(color.FgMagenta).SprintFunc()

type CrawlInfo struct {
	Urls   []string `json:"urls"`
	Length int      `json:"length"`
	Time   string   `json:"time"`
}

func helpPanel() {
	fmt.Println(`Usage of gr-crawl:
  INPUT:
    -u, -url string       target url to crawl to gather urls (i.e. https://example.com)
    -l, -list string      file containing a list of urls to crawl (one url per line)

  OUTPUT:
    -o, -output string          file to write urls into
    -oj, -output-json string    file to write urls into (JSON format)
    -oc, -output-csv string     file to write urls into (CSV format)

  CRAWLING:
    -d, -depth int      depth to crawl (default=2)
    -j, -js             only crawl JS endpoints (default=disabled)
    -path               only crawl urls inside each url path (default=disabled)
    -s, -subs           also crawl subdomains (default=disabled)
    -smart              enable "smart mode" to exclude useless stuff from results

  CONFIG:
    -w, -workers int      number of concurrent workers (default=10)
    -p, -proxy string     proxy to send requests through (i.e. http://127.0.0.1:8080)
    -t, -timeout int      millisecond to wait before each request timeout (default=8000)
    -c, -color            print colors on output
    -q, -quiet            print neither banner nor logging, only print output

  DEBUG:
    -version      show go-recon version
    -h, -help     print help panel
  
Examples:
    gr-crawl -u https://example.com -o endpoints.txt -c
    gr-crawl -l urls.txt -d 3 -w 15
    gr-crawl -u https://example.com -js -oj urls.json
    cat urls.txt | gr-crawl -path -q
    `)
}

var js bool    // define it globally so it's accessible from printResult() function
var smart bool // define it globally so it's accessible from printResult() function

var smart_exts []string = []string{"png", "jpg", "jpeg"}

var smart_filenames []string = []string{"google_tag.script.js", "forms2.min.js", "highcharts.js", "highcharts-more.js", "uc.js", "v2.js", "jquery.min.js"}

var smart_prefixes []string = []string{"wp-runtime-", "jquery."}

var csv_urls [][]string

// nolint: gocyclo
func main() {
	var url string
	var list string
	var depth int
	var workers int
	var proxy string
	var output string
	var json_output string
	var csv_output string
	var subs bool
	var inside_path bool
	var use_color bool
	var timeout int
	var quiet bool
	var version bool
	var help bool
	var stdin bool

	flag.StringVar(&url, "u", "", "")
	flag.StringVar(&url, "url", "", "")
	flag.StringVar(&list, "l", "", "")
	flag.StringVar(&list, "list", "", "")
	flag.IntVar(&depth, "d", 2, "")
	flag.IntVar(&depth, "depth", 2, "")
	flag.BoolVar(&smart, "smart", false, "")
	flag.BoolVar(&inside_path, "path", false, "")
	flag.IntVar(&workers, "w", 10, "")
	flag.IntVar(&workers, "workers", 10, "")
	flag.StringVar(&proxy, "p", "", "")
	flag.StringVar(&proxy, "proxy", "", "")
	flag.StringVar(&output, "o", "", "")
	flag.StringVar(&output, "output", "", "")
	flag.StringVar(&json_output, "oj", "", "")
	flag.StringVar(&json_output, "output-json", "", "")
	flag.StringVar(&csv_output, "oc", "", "")
	flag.StringVar(&csv_output, "output-csv", "", "")
	flag.BoolVar(&js, "j", false, "")
	flag.BoolVar(&js, "js", false, "")
	flag.BoolVar(&subs, "s", false, "")
	flag.BoolVar(&subs, "subs", false, "")
	flag.BoolVar(&quiet, "q", false, "")
	flag.BoolVar(&quiet, "quiet", false, "")
	flag.BoolVar(&use_color, "c", false, "")
	flag.BoolVar(&use_color, "color", false, "")
	flag.IntVar(&timeout, "t", 8000, "")
	flag.IntVar(&timeout, "timeout", 8000, "")
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

	results := make(chan string)

	var counter int
	var urls []string

	// add url or urls to array
	if url != "" {
		urls = append(urls, url)

	} else if list != "" {
		f, err := os.Open(list)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			urls = append(urls, scanner.Text())
		}

	} else if stdin {
		f := os.Stdin

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			urls = append(urls, scanner.Text())
		}
	}

	if !quiet {
		core.Warning("Use with caution.", use_color)
		core.Magenta("Concurrent workers: "+strconv.Itoa(workers), use_color)
		if proxy != "" {
			core.Magenta("Proxy: "+proxy, use_color)
		}
		core.Magenta("Starting crawling process on provided url(s)...\n", use_color)
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
				if strings.Contains(abs_link, url) || !inside_path {
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
				Proxy:           http.ProxyFromEnvironment,
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
			if res == u { // if current url is present in array, skip this url
				found = 1
				break
			}
		}

		if found == 1 {
			continue
		}

		unique_urls = append(unique_urls, res) // append url to unique urls array
		csv_urls = append(csv_urls, []string{res})

		counter += 1
		fmt.Println(res)

		if output != "" {
			_, err = txt_out.WriteString(res + "\n")
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	if json_output != "" {
		json_crawl := CrawlInfo{
			Urls:   unique_urls,
			Length: len(unique_urls),
			Time:   core.TimerDiff(t1).String(),
		}

		json_body, err := json.Marshal(json_crawl)
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

		headers := []string{"urls"}
		writer.Write(headers)

		for _, row := range csv_urls {
			writer.Write(row)
		}
	}

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

			core.Green(strconv.Itoa(counter)+" urls found", use_color)
		} else {
			core.Red("No urls gathered", use_color)
		}

		if use_color {
			fmt.Println("["+green("+")+"] Elapsed time:", green(core.TimerDiff(t1)))
		} else {
			fmt.Println("[+] Elapsed time:", core.TimerDiff(t1))
		}
	}
}

// nolint: gocyclo
func printResult(link string, sourceName string, results chan string, e *colly.HTMLElement) {
	var result string

	if strings.HasPrefix(link, "////") {
		result = "https:" + strings.Replace(link, "////", "//", 1)
	} else if strings.HasPrefix(link, "//") {
		result = "https:" + link
	} else {
		result = e.Request.AbsoluteURL(link)
	}

	if js {
		ext, err := getUrlExtension(result)
		if err != nil && err.Error() != "couldn't find a period to indicate a file extension" {
			log.Fatal(err)
		}

		if ext != "js" {
			return
		}

		r := regexp.MustCompile(`[(\w./:)]*js`)
		matches := r.FindAllString(result, -1)
		for _, j := range matches {

			if strings.HasPrefix(j, "//") {
				results <- "https:" + j
			} else if strings.HasPrefix(j, "/") {
				results <- j
			}
		}
	}

	if smart {
		// parse url to get url filename and compare it
		u, err := nurl.Parse(result)
		if err != nil {
			log.Fatal(err)
		}
		x, _ := nurl.QueryUnescape(u.EscapedPath())
		url_filename := filepath.Base(x)

		pos := strings.LastIndex(u.Path, ".")
		if pos != -1 {
			current_ext := u.Path[pos+1 : len(u.Path)]

			for _, f := range smart_filenames {
				if url_filename == f {
					return
				}
			}

			for _, e := range smart_exts {
				if e == current_ext {
					return
				}
			}

			for _, prefix := range smart_prefixes {
				if strings.HasPrefix(url_filename, prefix) {
					return
				}
			}
		}
	}

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

func getUrlExtension(rawUrl string) (string, error) {
	u, err := nurl.Parse(rawUrl)
	if err != nil {
		return "", err
	}

	pos := strings.LastIndex(u.Path, ".")
	if pos == -1 {
		return "", errors.New("couldn't find a period to indicate a file extension")
	}

	return u.Path[pos+1 : len(u.Path)], nil
}
