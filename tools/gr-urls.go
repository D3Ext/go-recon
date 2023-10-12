package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"

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
	fmt.Println(`Usage of gr-urls:
  INPUT:
    -d, -domain string      domain to retrieve a bunch of urls from different sources (i.e. example.com)

  OUTPUT:
    -o, -output string          file to write urls into
    -oj, -output-json string    file to write urls into (JSON format)

  CONFIG:
    -nr, -no-recursive    retrieve urls just for given domain, without subdomains (default=disabled)
    -p, -proxy string     proxy to send requests through (i.e. http://127.0.0.1:8080)
    -w, -workers int      number of concurrent workers (default=2)
    -t, -timeout int      milliseconds to wait before each request timeout (default=15000)
    -c, -color            print colors on output
    -q, -quiet            print neither banner nor logging, only print output

  DEBUG:
    -version      show go-recon version
    -h, -help     print help panel

Examples:
    gr-urls -d example.com -o urls.txt
    gr-urls -d example.com -nr 
    echo "example.com" | gr-urls
    `)
}

var found_urls []string
var current_url string
var counter int

// nolint: gocyclo
func main() {
	var domain string
	var workers int
	var proxy string
	var output string
	var json_output string
	var recursive bool
	var timeout int
	var quiet bool
	var use_color bool
	var version bool
	var help bool
	var stdin bool

	flag.StringVar(&domain, "d", "", "")
	flag.StringVar(&domain, "domain", "", "")
	flag.IntVar(&workers, "w", 2, "")
	flag.IntVar(&workers, "workers", 2, "")
	flag.StringVar(&proxy, "p", "", "")
	flag.StringVar(&proxy, "proxy", "", "")
	flag.StringVar(&output, "o", "", "")
	flag.StringVar(&output, "output", "", "")
	flag.StringVar(&json_output, "oj", "", "")
	flag.StringVar(&json_output, "output-json", "", "")
	flag.BoolVar(&recursive, "nr", false, "")
	flag.BoolVar(&recursive, "no-recursive", false, "")
	flag.IntVar(&timeout, "t", 15000, "")
	flag.IntVar(&timeout, "timeout", 15000, "")
	flag.BoolVar(&quiet, "q", false, "")
	flag.BoolVar(&quiet, "quiet", false, "")
	flag.BoolVar(&use_color, "c", false, "")
	flag.BoolVar(&use_color, "color", false, "")
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

	if help { // Print help panel
		helpPanel()
		os.Exit(0)
	}

	fi, err := os.Stdin.Stat()
	if err != nil {
		log.Fatal(err)
	}

	if fi.Mode()&os.ModeNamedPipe == 0 {
		stdin = false // stdin is empty
	} else {
		stdin = true // stdin has value
	}

	if stdin {
		// Read STDIN
		stdin, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}

		if string(stdin) != "" {
			domain = strings.TrimSuffix(string(stdin), "\n")
		}
	}

	if (domain == "") && (!stdin) {
		helpPanel()
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

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		if json_output != "" {
			json_urls := UrlsInfo{
				Urls:   found_urls,
				Length: len(found_urls),
				Time:   core.TimerDiff(t1).String(),
			}

			json_out, err := os.Create(json_output)
			if err != nil {
				log.Fatal(err)
			}

			json_body, err := json.Marshal(json_urls)
			if err != nil {
				log.Fatal(err)
			}

			_, err = json_out.WriteString(string(json_body))
			if err != nil {
				log.Fatal(err)
			}
		}
		os.Exit(0)
	}()

	if !quiet {
		core.Warning("Use with caution.", use_color)
		core.Magenta("Gathering "+domain+" urls...\n", use_color)
	}

	if recursive {
		recursive = false
	} else {
		recursive = true
	}

	client := core.CreateHttpClient(timeout)

	results := make(chan string)

	go func() {
		for res := range results {
			fmt.Println(res)
			counter += 1
			found_urls = append(found_urls, res)

			if output != "" {
				_, err = txt_out.WriteString(res + "\n")
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}()

	err = core.GetAllUrls(domain, results, client, recursive)
	if err != nil {
		log.Fatal(err)
	}

	if json_output != "" {
		json_urls := UrlsInfo{
			Urls:   found_urls,
			Length: len(found_urls),
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

	if !quiet {
		fmt.Println()
		if counter >= 1 {
			if output != "" {
				core.Green("Urls written to "+output+" (TXT)", use_color)
			}

			if json_output != "" {
				core.Green("Urls written to "+json_output+" (JSON)", use_color)
			}
		}

		if use_color {
			fmt.Println("["+green("+")+"]", cyan(counter), "urls obtained")
		} else {
			fmt.Println("[+]", counter, "urls obtained")
		}

		if use_color {
			fmt.Println("["+green("+")+"] Elapsed time:", green(core.TimerDiff(t1).String()))
		} else {
			fmt.Println("[+] Elapsed time:", core.TimerDiff(t1).String())
		}
	}
}
