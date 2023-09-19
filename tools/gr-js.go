package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"log"
	"os"
	"sync"

	"github.com/D3Ext/go-recon/core"
)

var red func(a ...interface{}) string = color.New(color.FgRed).SprintFunc()
var cyan func(a ...interface{}) string = color.New(color.FgCyan).SprintFunc()
var green func(a ...interface{}) string = color.New(color.FgGreen).SprintFunc()
var magenta func(a ...interface{}) string = color.New(color.FgMagenta).SprintFunc()

type JsInfo struct {
	Endpoints []string `json:"endpoints"`
	Length    int      `json:"length"`
	Time      string   `json:"time"`
}

func helpPanel() {
	fmt.Println(`Usage of gr-js:
    -l)       file containing a list of urls to find their endpoints (one url per line)
    -w)       number of concurrent workers (default=15)
    -o)       file to write JS endpoints into
    -oj)      file to write JS endpoints into (JSON format)
    -p)       proxy to send requests through (i.e. http://127.0.0.1:8080)
    -a)       user agent to include on requests (default=none)
    -c)       print colors on output
    -t)       milliseconds to wait before each request timeout (default=5000)
    -q)       don't print banner, only output
    -h)       print help panel
  
Examples:
    gr-js -l urls.txt -o endpoints.txt
    gr-js -l urls.txt -w 10 -t 6000
    cat urls.txt | gr-js
    `)
}

func main() {
	var list string
	var workers int
	var output string
	var json_output string
	var proxy string
	var stdin bool
	var user_agent string
	var timeout int
	var quiet bool
	var use_color bool
	var help bool

	flag.StringVar(&list, "l", "", "file containing a list of urls to find their endpoints (one url per line)")
	flag.IntVar(&workers, "w", 15, "number of concurrent workers")
	flag.StringVar(&user_agent, "a", "", "user agent to include on requests")
	flag.StringVar(&output, "o", "", "file to write JS endpoints into")
	flag.StringVar(&json_output, "oj", "", "file to write JS endpoints into (JSON format)")
	flag.StringVar(&proxy, "p", "", "proxy to send requests through (i.e. http://127.0.0.1:8080)")
	flag.IntVar(&timeout, "t", 5000, "milliseconds to wait before each request timeout")
	flag.BoolVar(&quiet, "q", false, "don't print banner, only output")
	flag.BoolVar(&use_color, "c", false, "print colors on output")
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

	var err error

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

	if (list == "") && (!stdin) { // check if required arguments were given
		helpPanel()
		os.Exit(0)
	}

	var out_f *os.File
	if output != "" { // create output file if it was provided
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

	var counter int

	if !quiet {
		core.Magenta("Fetching js endpoints from given urls...\n", use_color)
	}

	if (list != "") || (stdin) {

		var f *os.File

		if list != "" { // get file descriptor from file or stdin
			f, err = os.Open(list)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

		} else if stdin {
			f = os.Stdin
		}

		urls_c := make(chan string)  // create channel to receive urls
		results := make(chan string) // create channel to receive endpoints
		var wg sync.WaitGroup        // create work group

		for i := 0; i < workers; i++ { // create n workers
			wg.Add(1)

			go func() {
				core.FetchEndpoints(
					urls_c,
					results,
					"Mozilla/5.0 (X11; Linux x86_64; rv:78.0) Gecko/20100101 Firefox/78.0",
					proxy,
					timeout,
				) // fetch js endpoints

				wg.Done()
			}()
		}

		var out sync.WaitGroup // create output wait group
		out.Add(1)             // create one worker

		var endpoints []string

		go func() {
			for res := range results { // receive endpoints from channel and print them
				fmt.Println(res)
				endpoints = append(endpoints, res)
				counter += 1

				if output != "" { // write endpoint to output file if it was provided
					_, err = out_f.WriteString(res + "\n")
					if err != nil {
						log.Fatal(err)
					}
				}
			}

			out.Done() // finish out worker since results channel'll be closed when all workers finish so the loop also finish and out.Done() is executed
		}()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() { // iterate over file or stdin
			line := scanner.Text()

			if line != "" {
				urls_c <- line // sent urls through chhanel
			}
		}

		close(urls_c)  // close urls channel
		wg.Wait()      // wait for all workers
		close(results) // close results channel once workers have finished
		out.Wait()     // and finally wait for single output worker

		if json_output != "" {
			json_endpoints := JsInfo{
				Endpoints: endpoints,
				Length:    counter,
				Time:      core.TimerDiff(t1).String(),
			}

			json_body, err := json.Marshal(json_endpoints)
			if err != nil {
				log.Fatal(err)
			}

			_, err = out_f.WriteString(string(json_body))
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	// finally add some logging to aid users
	if !quiet {
		if counter > 1 {
			if use_color {
				fmt.Println("\n["+green("+")+"]", cyan(counter), "endpoints found")
				if output != "" {
					core.Green("Endpoints written to "+output, use_color)
				} else if json_output != "" {
					core.Green("Endpoints written to "+json_output, use_color)
				}
				fmt.Println("["+green("+")+"] Elapsed time:", green(core.TimerDiff(t1)))
			} else {
				fmt.Println("\n[+]", counter, "endpoints found")
				if output != "" {
					core.Green("Endpoints written to "+output, use_color)
				} else if json_output != "" {
					core.Green("Endpoints written to "+json_output, use_color)
				}
				fmt.Println("[+] Elapsed time:", core.TimerDiff(t1))
			}

		} else {
			if use_color {
				fmt.Println("\n["+green("+")+"] Elapsed time:", green(core.TimerDiff(t1)))
			} else {
				fmt.Println("\n[+] Elapsed time:", core.TimerDiff(t1))
			}
		}
	}
}
