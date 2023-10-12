package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/D3Ext/go-recon/core"
)

var red func(a ...interface{}) string = color.New(color.FgRed).SprintFunc()
var cyan func(a ...interface{}) string = color.New(color.FgCyan).SprintFunc()
var green func(a ...interface{}) string = color.New(color.FgGreen).SprintFunc()
var magenta func(a ...interface{}) string = color.New(color.FgMagenta).SprintFunc()

type UrlsInfo struct {
	Urls   []string `json:"urls"`
	Length int      `json:"length"`
	Time   string   `json:"time"`
}

func helpPanel() {
	fmt.Println(`Usage of gr-replace:
  INPUT:
    -l, -list string      file containing a list of urls to replace specific strings on them (one url per line)

  OUTPUT:
    -o, -output string          file to write modified urls into
    -oj, -output-json string    file to write modified urls into (JSON format)

  CONFIG:
    -k, -keyword string     keyword to replace in urls with supplied value (i.e. FUZZ)
    -p, -param string       parameter name to replace its value with keyword from -k parameter (i.e. id)
    -v, -value string       value to replace in urls with keyword from -k parameter (i.e. <script>alert('XSS')</script> )
    -hide                   don't print unmodified urls
    -c, -color              print colors on output
    -q, -quiet              print neither banner nor logging, only print output

  DEBUG:
    -version        show go-recon version
    -h, -help       print help panel

Examples:
    gr-replace -l urls.txt -k FUZZ -v "<script>alert('XSS')</script>" -o new_urls.txt
    gr-replace -l urls.txt -p id -v "' or 1=1-- -" -c
    cat urls.txt | gr-replace -hide
    `)
}

// nolint: gocyclo
func main() {
	var list string
	var keyword string
	var param string
	var value string
	var hide bool
	var output string
	var json_output string
	var quiet bool
	var use_color bool
	var version bool
	var help bool
	var stdin bool

	flag.StringVar(&list, "l", "", "")
	flag.StringVar(&list, "list", "", "")
	flag.StringVar(&keyword, "k", "", "")
	flag.StringVar(&keyword, "keyword", "", "")
	flag.StringVar(&param, "p", "", "")
	flag.StringVar(&param, "param", "", "")
	flag.StringVar(&value, "v", "", "")
	flag.StringVar(&value, "value", "", "")
	flag.BoolVar(&hide, "hide", false, "")
	flag.StringVar(&output, "o", "", "")
	flag.StringVar(&output, "output", "", "")
	flag.StringVar(&json_output, "oj", "", "")
	flag.StringVar(&json_output, "output-json", "", "")
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

	if (value == "") || (param == "" && keyword == "") || (param != "" && keyword != "") {
		helpPanel()
		core.Red("You need to provide a valid keyword (-k) or param value (-p) to replace with supplied value (-v)", use_color)
		os.Exit(0)
	}

	var txt_out *os.File
	if output != "" { // create output file if it was provided
		txt_out, err = os.Create(output)
		if err != nil {
			log.Fatal(err)
		}
	}

	var counter int
	var replaced_urls []string

	if !quiet {
		core.Magenta("Value: "+value, use_color)
		core.Magenta("Processing urls and replacing values...\n", use_color)
	}

	if (list != "") || (stdin) {

		var f *os.File

		if list != "" {
			f, err = os.Open(list)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

		} else {
			f = os.Stdin
		}

		var line string
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line_pre_change := scanner.Text()

			u, _ := url.Parse(line_pre_change) // parse each url
			if u == nil {
				continue
			}

			if (keyword != "") && (value != "") {
				line = strings.ReplaceAll(line_pre_change, keyword, value)

			} else if (param != "") && (value != "") {
				for k := range u.Query() { // iterate over parameters and their values
					if k == param { // check param name
						line = strings.ReplaceAll(line_pre_change, u.Query().Get(k), value) // replace param value with provided value to replace it
						break
					} else {
						line = line_pre_change
					}
				}
			}

			if (line_pre_change != line) && (hide) {
				fmt.Println(line)
				replaced_urls = append(replaced_urls, line)
				counter += 1

			} else if !hide {

				if line != "" {
					fmt.Println(line)
					replaced_urls = append(replaced_urls, line)
				} else {
					fmt.Println(line_pre_change)
					replaced_urls = append(replaced_urls, line_pre_change)
				}

				counter += 1
			}

			if output != "" {
				_, err = txt_out.WriteString(line + "\n")
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}

	if json_output != "" {
		json_urls := UrlsInfo{
			Urls:   replaced_urls,
			Length: len(replaced_urls),
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

	// finally add some logging to aid users
	if !quiet {
		if counter >= 1 {
			fmt.Println()
			if output != "" {
				core.Green("Results written to "+output+" (TXT)", use_color)
			}

			if json_output != "" {
				core.Green("Results written to "+json_output+" (JSON)", use_color)
			}

			core.Green(strconv.Itoa(counter)+" lines processed", use_color)
		}

		if use_color {
			fmt.Println("["+green("+")+"] Elapsed time:", green(core.TimerDiff(t1)))
		} else {
			fmt.Println("[+] Elapsed time:", core.TimerDiff(t1))
		}
	}
}
