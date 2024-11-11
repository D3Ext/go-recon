package main

import (
	"bufio"
  "encoding/csv"
	"encoding/json"
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

type SubdomainsInfo struct {
	Subdomains []string `json:"subdomains"`
	Length     int      `json:"length"`
	Time       string   `json:"time"`
}

func listProviders(quiet bool) {
	if !quiet {
		fmt.Println(`[*] Available providers:
    - alienvault
    - anubis
    - commoncrawl
    - crt
    - digitorus
    - hackertarget
    - rapiddns
    - wayback
  
    Providers used by default: [alienvault, crt, hackertarget, rapiddns, digitorus]`)
	} else {
		fmt.Println("alienvault\nanubis\ncommoncrawl\ncrt\ndigitorus\nhackertarget\nrapiddns\nwayback")
	}
}

func helpPanel() {
	fmt.Println(`Usage of gr-subdomains:
  INPUT:
    -d, -domain string      domain to find its subdomains (i.e. example.com)
    -l, -list string        file containing a list of domains to find their subdomains (one domain per line)

  OUTPUT:
    -o, -output string          file to write subdomains into (TXT format)
    -oj, -output-json string    file to write subdomains into (JSON format)
    -oc, -output-csv string     file to write subdomains into (CSV format)

  PROVIDERS:
    -all                      use all available providers to discover subdomains (slower than default)
    -p, -providers string[]   providers to use for subdomain discovery (separated by comma)
    -lp, -list-providers      list available providers

  CONFIG:
    -proxy string         proxy to send requests through (i.e. http://127.0.0.1:8080)
    -t, -timeout int      milliseconds to wait before each request timeout (default=5000)
    -c, -color            print colors on output
    -q, -quiet            print neither banner nor logging, only print output

  DEBUG:
    -version      show go-recon version
    -h, -help     print help panel
  
Examples:
    gr-subdomains -d example.com -o subdomains.txt -c
    gr-subdomains -l domains.txt -p crt,hackertarget -t 8000
    cat domain.txt | gr-subdomains -all -q
    cat domain.txt | gr-subdomains -p anubis -oj subdomains.json -c
    `)
}

// nolint: gocyclo
func main() {
	var domain string
	var list string
	var all bool
	var providers_str string
	var list_providers bool
	var output string
	var json_output string
  var csv_output string
	var proxy string
	var stdin bool
	var timeout int
	var quiet bool
	var use_color bool
	var version bool
	var help bool

	flag.StringVar(&domain, "d", "", "")
	flag.StringVar(&domain, "domain", "", "")
	flag.StringVar(&list, "l", "", "")
	flag.StringVar(&list, "list", "", "")
	flag.BoolVar(&all, "all", false, "")
	flag.StringVar(&providers_str, "p", "", "")
	flag.StringVar(&providers_str, "providers", "", "")
	flag.BoolVar(&list_providers, "lp", false, "")
	flag.BoolVar(&list_providers, "list-providers", false, "")
	flag.StringVar(&output, "o", "", "")
	flag.StringVar(&output, "output", "", "")
	flag.StringVar(&json_output, "oj", "", "")
	flag.StringVar(&json_output, "output-json", "", "")
  flag.StringVar(&csv_output, "oc", "", "")
  flag.StringVar(&csv_output, "output-csv", "", "")
	flag.StringVar(&proxy, "proxy", "", "")
	flag.IntVar(&timeout, "t", 5000, "")
	flag.IntVar(&timeout, "timeout", 5000, "")
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

	if list_providers {
		listProviders(quiet)
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

	if providers_str == "" { // default providers
		providers_str = "alienvault,crt,digitorus,hackertarget,rapiddns"
	} else if all { // all providers
		providers_str = "alienvault,anubis,commoncrawl,crt,digitorus,hackertarget,rapiddns,wayback"
	}

	providers := strings.Split(strings.ToLower(strings.ReplaceAll(providers_str, " ", "")), ",")

	// define variables which will be used to write subdomains to file
	var txt_out *os.File
	if output != "" {
		txt_out, err = os.Create(output)
		if err != nil {
			log.Fatal(err)
		}
	}

	var counter int
	var total_counter int
	var found_subdomains []string
  var csv_info [][]string

	client := core.CreateHttpClient(timeout)

	// Create channel for asynchronous jobs
	results := make(chan string)

	var wg sync.WaitGroup
	var out sync.WaitGroup
	out.Add(1)

	if !quiet {
		core.Warning("Use with caution", use_color)
	}

	if domain != "" { // if domain parameter has value enter here

		if !quiet {
			core.Magenta("Discovering "+domain+" subdomains...", use_color)
			core.Magenta(fmt.Sprintf("Selected providers: %s\n", providers), use_color)
		}

		wg.Add(1) // used to wait until main goroutine finish, that goroutine will generate 8 concurrent workers

		go func() { // goroutine used for printing results
			for subdomain := range results {
				//fmt.Println(subdomain)
				if checkSubdomain(subdomain, found_subdomains) {
					fmt.Println(subdomain)
					found_subdomains = append(found_subdomains, subdomain)

					if output != "" { // if output parameter has value, write subdomains to file
						_, err = txt_out.WriteString(subdomain + "\n")
						if err != nil {
							log.Fatal(err)
						}
					}

          csv_info = append(csv_info, []string{subdomain, domain})

					counter++
				}
			}

			out.Done()
		}()

		// run all providers enum
		go func() {
			defer wg.Done()

			//func GetAllSubdomains(dom string, c chan string, providers []string, client *http.Client) error
			core.GetSubdomains(domain, results, providers, client)
		}()

		wg.Wait()
		close(results)
		out.Wait()

	} else if (list != "") || (stdin) { // enter here if list parameter has value or if stdin has value

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

		var doms_length int
		var domains []string

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if (scanner.Text() != "") && (strings.Contains(scanner.Text(), ".")) {
				doms_length += 1
				domains = append(domains, scanner.Text())
			}
		}

		if !quiet {
			if use_color {
				fmt.Println("["+magenta("*")+"] Discovering subdomains for", green(doms_length), "domains")
			} else {
				fmt.Println("[*] Discovering subdomains for", doms_length, "domains")
			}
			core.Magenta(fmt.Sprintf("Selected providers: %s", providers), use_color)
		}

		// Create channel for asynchronous jobs
		results := make(chan string)

		go func() { // goroutine used for printing results
			for subdomain := range results {
				if checkSubdomain(subdomain, found_subdomains) {
					fmt.Println(subdomain)
					found_subdomains = append(found_subdomains, subdomain)

					if output != "" { // if output parameter has value, write subdomains to file
						_, err = txt_out.WriteString(subdomain + "\n")
						if err != nil {
							log.Fatal(err)
						}
					}

          csv_info = append(csv_info, []string{subdomain, domain})

					total_counter += 1
					counter++
				}
			}

			out.Done()
		}()

		for _, dom := range domains {
			wg.Add(1)

			if !quiet {
				core.Magenta("Enumerating "+dom+" subdomains", use_color)
			}

			core.GetSubdomains(dom, results, providers, client)

			if !quiet {
				core.Green(strconv.Itoa(counter)+" subdomains found for "+dom+"\n", use_color)
			}

			counter = 0
			wg.Done()
		}

		wg.Wait()
		close(results)
		out.Wait()
	}

	if json_output != "" {
		json_subs := SubdomainsInfo{
			Subdomains: found_subdomains,
			Length:     counter,
			Time:       core.TimerDiff(t1).String(),
		}

		json_body, err := json.Marshal(json_subs)
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

		headers := []string{"subdomain", "domain"}

		writer.Write(headers)
		for _, row := range csv_info {
			writer.Write(row)
		}
	}

	// Finally some logging to aid users
	if !quiet {
		if total_counter >= 1 {
			if use_color {
				fmt.Println("["+green("+")+"]", cyan(total_counter), "subdomains found in total")
			} else {
				fmt.Println("[+]", total_counter, "subdomains found in total")
			}
		} else if counter >= 1 {
			if use_color {
				fmt.Println("\n["+green("+")+"]", cyan(counter), "subdomains found")
			} else {
				fmt.Println("\n[+]", counter, "subdomains found")
			}
		}

		if total_counter >= 1 || counter >= 1 { // Check if at least one url was vulnerable to open redirect
			if output != "" {
				core.Green("Subdomains written to "+output+" (TXT)", use_color)
			}

      if json_output != "" {
				core.Green("Subdomains written to "+json_output+" (JSON)", use_color)
			}

      if csv_output != "" {
        core.Green("Subdomains written to "+csv_output+" (CSV)", use_color)
      }
		}

		if use_color {
			fmt.Println("["+green("+")+"] Elapsed time:", green(core.TimerDiff(t1)))
		} else {
			fmt.Println("[+] Elapsed time:", core.TimerDiff(t1))
		}
	}
}

func checkSubdomain(str string, slice []string) bool {
	for _, s := range slice {
		if strings.ToLower(str) == strings.ToLower(s) {
			//fmt.Println(strings.ToLower(str), "-", strings.ToLower(s))
			return false
		}
	}

	return true
}
