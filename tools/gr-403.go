package main

import (
	"bufio"
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

var counter int

func check(url, word string, timeout int, use_color bool) error {
	url = strings.TrimSuffix(url, "/")
	bypasses, status_codes, err := core.Check403(url, word, timeout)
	if err != nil {
		return err
	}

	for i, b := range bypasses {
		if use_color {
			if status_codes[i] == 200 {
				fmt.Println("[" + magenta("*") + "] " + green("200") + " - " + b)
				counter += 1
			} else if (status_codes[i] == 302) || (status_codes[i] == 301) {
				fmt.Println("[" + magenta("*") + "] " + cyan("302") + " - " + b)
				counter += 1
			} else if status_codes[i] == 404 {
				fmt.Println("[" + magenta("*") + "] " + red("404") + " - " + b)
			} else {
				fmt.Println("[" + magenta("*") + "] " + yellow(strconv.Itoa(status_codes[i])) + " - " + b)
			}

		} else {
			fmt.Println("[*] " + strconv.Itoa(status_codes[i]) + " - " + b)
		}
	}

	return nil
}

func helpPanel() {
	fmt.Println(`Usage of gr-403:
    -u)       url to find potential 403 bypasses (i.e. https://example.com)
    -l)       file containing a list of urls to find potential bypasses (one url per line)
    -k)       keyword to test in url and headers (default="secret")
    -w)       number of concurrent workers (default=10)
    -t)       milliseconds to wait before each request timeout (default=5000)
    -c)       use color on output (recommended)
    -q)       don't print banner, only output
    -h)       print help panel
  
Examples:
    gr-403 -u https://example.com -c
    gr-403 -u https://example.com -k test
    gr-403 -l urls.txt -w 15 -t 4000
    cat urls.txt | gr-403
    `)
}

func main() {
	var url string
	var list string
	var word string
	var workers int
	var timeout int
	var use_color bool
	var quiet bool
	var help bool
	var stdin bool

	flag.StringVar(&url, "u", "", "url to find potential 403 bypasses (i.e. https://domain.com)")
	flag.StringVar(&list, "l", "", "file containing a list of urls to find potential bypasses (one url per line)")
	flag.StringVar(&word, "k", "secret", "keyword to test in url and headers")
	flag.IntVar(&workers, "w", 10, "number of concurrent workers")
	flag.IntVar(&timeout, "t", 5000, "milliseconds to wait before each request timeout (default=5000)")
	flag.BoolVar(&use_color, "c", false, "print colors on output (recommended)")
	flag.BoolVar(&quiet, "q", false, "don't print banner, only output")
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

	if !quiet {
		core.Green("Finding possible 403 bypasses...", use_color)
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

			//fmt.Println(line)
			if line != "" {
				urls_c <- line // send url through channel
			}
		}

		close(urls_c)
		wg.Wait()
	}

	// finally some logging to aid users
	if !quiet {
		if counter >= 1 {
			if use_color {
				fmt.Println("\n["+green("+")+"]", cyan(counter), "bypasses found!")
			} else {
				fmt.Println("\n[+]", counter, "bypasses found!")
			}
		}

		if use_color {
			if counter >= 1 {
				fmt.Println("["+green("+")+"] Elapsed time:", green(core.TimerDiff(t1)))
			} else {
				fmt.Println("\n["+green("+")+"] Elapsed time:", green(core.TimerDiff(t1)))
			}
		} else {
			if counter >= 1 {
				fmt.Println("[+] Elapsed time:", core.TimerDiff(t1))
			} else {
				fmt.Println("\n[+] Elapsed time:", core.TimerDiff(t1))
			}
		}
	}
}
