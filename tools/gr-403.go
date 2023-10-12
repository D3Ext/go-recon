package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"log"
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

func check(url, word string, timeout int, use_color bool) error {
	url = strings.TrimSuffix(url, "/")
	bypasses, status_codes, err := core.Check403(url, word, timeout)
	if err != nil {
		return err
	}

	for i, b := range bypasses {
		if skip {
			if status_codes[i] == 403 {
				continue
			}
		}

		if use_color {
			if status_codes[i] == 200 {
				fmt.Println(green("200") + " - " + b)
			} else if (status_codes[i] == 302) || (status_codes[i] == 301) {
				fmt.Println(cyan("302") + " - " + b)
			} else if status_codes[i] == 404 {
				fmt.Println(red("404") + " - " + b)
			} else {
				fmt.Println(yellow(strconv.Itoa(status_codes[i])) + " - " + b)
			}

		} else {
			fmt.Println(strconv.Itoa(status_codes[i]) + " - " + b)
		}

		if status_codes[i] != 403 {
			csv_info = append(csv_info, []string{b, strconv.Itoa(status_codes[i])})
			counter += 1
		}
	}

	return nil
}

func helpPanel() {
	fmt.Println(`Usage of gr-403:
  INPUT:
    -u, -url string       url to find potential 403 bypasses (i.e. https://example.com)
    -l, -list string      file containing a list of urls to find potential bypasses (one url per line)

  OUTPUT:
    -o, -output string    file to write vulnerable urls into (CSV format)

  CONFIG:
    -s, -skip               don't show urls that return 403 status code
    -k, -keyword string     keyword to test in url and headers (default="secret")
    -p, -proxy string       proxy to send requests through (i.e. http://127.0.0.1:8080)
    -w, -workers int        number of concurrent workers (default=10)
    -t, -timeout int        milliseconds to wait before each request timeout (default=5000)
    -c, -color              use color on output (recommended)
    -q, -quiet              don't print banner, only output

  DEBUG:
    -version        show go-recon version
    -h, -help       print help panel
  
Examples:
    gr-403 -u https://example.com -c
    gr-403 -u https://example.com -k test
    gr-403 -l urls.txt -w 15 -t 4000
    cat urls.txt | gr-403
    `)
}

var counter int
var skip bool
var csv_info [][]string

// nolint: gocyclo
func main() {
	var url string
	var list string
	var csv_output string
	var word string
	var proxy string
	var workers int
	var timeout int
	var use_color bool
	var quiet bool
	var version bool
	var help bool
	var stdin bool

	flag.StringVar(&url, "u", "", "")
	flag.StringVar(&url, "url", "", "")
	flag.StringVar(&list, "l", "", "")
	flag.StringVar(&list, "list", "", "")
	flag.StringVar(&csv_output, "o", "", "")
	flag.StringVar(&csv_output, "output", "", "")
	flag.BoolVar(&skip, "s", false, "")
	flag.BoolVar(&skip, "skip", false, "")
	flag.StringVar(&word, "k", "secret", "")
	flag.StringVar(&word, "keyword", "secret", "")
	flag.StringVar(&proxy, "p", "", "")
	flag.StringVar(&proxy, "proxy", "", "")
	flag.IntVar(&workers, "w", 10, "")
	flag.IntVar(&workers, "workers", 10, "")
	flag.IntVar(&timeout, "t", 5000, "")
	flag.IntVar(&timeout, "timeout", 5000, "")
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
		stdin = false
	} else {
		stdin = true
	}

	if (url == "") && (list == "") && (!stdin) {
		helpPanel()
		os.Exit(0)
	}

	if (url != "") && (list != "") {
		helpPanel()
		core.Red("You can't use (-d) and (-l) at same time", use_color)
		os.Exit(0)
	}

	if proxy != "" {
		os.Setenv("HTTP_PROXY", proxy)
		os.Setenv("HTTPS_PROXY", proxy)
	}

	if !quiet {
		core.Warning("Use with caution.", use_color)
		core.Magenta("Concurrent workers: "+strconv.Itoa(workers), use_color)
		if proxy != "" {
			core.Magenta("Proxy: "+proxy, use_color)
		}
		core.Magenta("Finding possible 403 bypasses...\n", use_color)
	}

	if url != "" {
		check(url, word, timeout, use_color)

	} else if (list != "") || (stdin) {

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

		urls_c := make(chan string)
		var wg sync.WaitGroup

		for i := 0; i < workers; i++ {
			wg.Add(1)

			go func() {
				for u := range urls_c {
					err = check(u, word, timeout, use_color)
					if err != nil {
						log.Fatal(err)
					}
				}

				wg.Done()
			}()
		}

		scanner := bufio.NewScanner(f)
		for scanner.Scan() { // iterate over every single line
			line := scanner.Text()

			if line != "" {
				urls_c <- line // send url through channel
			}
		}

		close(urls_c)
		wg.Wait()
	}

	if csv_output != "" {
		csv_out, err := os.Create(csv_output)
		if err != nil {
			log.Fatal(err)
		}

		writer := csv.NewWriter(csv_out)
		defer writer.Flush()

		headers := []string{"urls", "status_codes"}
		writer.Write(headers)

		for _, row := range csv_info {
			writer.Write(row)
		}
	}

	// finally some logging to aid users
	if !quiet {
		if counter >= 1 {
			if use_color {
				fmt.Println("\n["+green("+")+"]", cyan(counter), "bypasses found!")
			} else {
				fmt.Println("\n[+]", counter, "bypasses found!")
			}

			if csv_output != "" {
				core.Green("Vulnerable urls written to "+csv_output+" (CSV)", use_color)
			}
		} else {
			core.Red("No vulnerable url found", use_color)
		}

		if use_color {
			fmt.Println("["+green("+")+"] Elapsed time:", green(core.TimerDiff(t1)))
		} else {
			fmt.Println("[+] Elapsed time:", core.TimerDiff(t1))
		}
	}
}
