package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"log"
	"os"
	"strings"

	"github.com/D3Ext/go-recon/core"
)

var red func(a ...interface{}) string = color.New(color.FgRed).SprintFunc()
var cyan func(a ...interface{}) string = color.New(color.FgCyan).SprintFunc()
var green func(a ...interface{}) string = color.New(color.FgGreen).SprintFunc()
var magenta func(a ...interface{}) string = color.New(color.FgMagenta).SprintFunc()
var yellow func(a ...interface{}) string = color.New(color.FgYellow).SprintFunc()

func printInfo(domain string, dns_info core.DnsInfo, color bool) {
	if color {
		fmt.Println("[" + magenta("+") + "] Domain: " + green(domain))
		fmt.Println(yellow("\tCNAME: ") + cyan(dns_info.CNAME))
		fmt.Println(yellow("\n\tTXT:"))
		for _, t := range dns_info.TXT {
			fmt.Println("\t  " + t)
		}

		fmt.Println(yellow("\n\tMX:"))
		for _, mx := range dns_info.MX {
			fmt.Println("\t  " + mx.Host)
		}

		fmt.Println(yellow("\n\tNS:"))
		for _, ns := range dns_info.NS {
			fmt.Println("\t  " + ns.Host)
		}

		fmt.Println(yellow("\n\tHosts:"))
		for _, host := range dns_info.Hosts {
			fmt.Println("\t  " + host)
		}

	} else {
		fmt.Println("[+] Domain: " + domain)
		fmt.Println("\tCNAME: " + dns_info.CNAME)
		fmt.Println("\n\tTXT:")
		for _, t := range dns_info.TXT {
			fmt.Println("\t  " + t)
		}

		fmt.Println("\n\tMX:")
		for _, mx := range dns_info.MX {
			fmt.Println("\t  " + mx.Host)
		}

		fmt.Println("\n\tNS:")
		for _, ns := range dns_info.NS {
			fmt.Println("\t  " + ns.Host)
		}

		fmt.Println("\n\tHosts:")
		for _, host := range dns_info.Hosts {
			fmt.Println("\t  " + host)
		}
	}
}

func helpPanel() {
	fmt.Println(`Usage of gr-dns:
    -d)       domain to find DNS information (i.e. example.com)
    -l)       file containing a list of domains to find their DNS info (one domain per line)
    -o)       file to write DNS info into (JSON format)
    -c)       print colors on output (recommended)
    -q)       don't print banner, only output
    -h)       print help panel

Examples:
    gr-dns -d example.com -c
    gr-dns -l domains.txt
    cat domains.txt | gr-dns
    `)
}

func main() {
	var domain string
	var list string
	var output string
	var stdin bool
	var quiet bool
	var use_color bool
	var help bool

	flag.StringVar(&domain, "d", "", "domain to find DNS information (i.e. example.com)")
	flag.StringVar(&list, "l", "", "file containing a list of domains to find their DNS info (one domain per line)")
	flag.StringVar(&output, "o", "", "file to write DNS info into (JSON format)")
	flag.BoolVar(&quiet, "q", false, "don't print banner, only output")
	flag.BoolVar(&use_color, "c", false, "print colors on output (recommended)")
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

	// check if at least one parameter was given
	if (domain == "") && (list == "") && (!stdin) {
		helpPanel()
		os.Exit(0)
	}

	// print error and exit in case domain and list were given
	if (domain != "") && (list != "") {
		helpPanel()
		core.Red("You can't use (-d) and (-l) at same time", use_color)
		os.Exit(0)
	}

	if domain != "" {
		if !strings.Contains(domain, ".") {
			core.Red("Invalid domain!", use_color)
			os.Exit(0)
		}

		dns_info, err := core.Dns(domain)
		if err != nil {
			log.Fatal(err)
		}

		printInfo(domain, dns_info, use_color)
		fmt.Println()

		if output != "" {
			json_body, err := json.Marshal(dns_info)
			if err != nil {
				log.Fatal(err)
			}

			out_f, err := os.Create(output)
			if err != nil {
				log.Fatal(err)
			}

			_, err = out_f.WriteString(string(json_body))
			if err != nil {
				log.Fatal(err)
			}
		}

	} else if (list != "") || (stdin) {

		var file *os.File
		var err error

		if list != "" {
			file, err = os.Open(list)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()

		} else {
			file = os.Stdin
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() { // Iterate over every single line
			if (scanner.Text() != "") && (strings.Contains(scanner.Text(), ".")) {
				dns_info, err := core.Dns(scanner.Text())
				if err != nil {
					log.Fatal(err)
				}

				printInfo(scanner.Text(), dns_info, use_color)
				fmt.Println()

			} else {
				core.Red("Invalid domain found! Skipping line...\n", use_color)
			}
		}
	}

	if !quiet {
		if output != "" {
			core.Green("DNS info written to "+output, use_color)
		}

		if use_color {
			fmt.Println("["+green("+")+"] Elapsed time:", green(core.TimerDiff(t1)))
		} else {
			fmt.Println("[+] Elapsed time:", core.TimerDiff(t1))
		}
	}
}
