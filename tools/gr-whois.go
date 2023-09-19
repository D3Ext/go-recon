package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/likexian/whois"
	wp "github.com/likexian/whois-parser"
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

func helpPanel() {
	fmt.Println(`Usage of gr-whois:
    -d)       domain to send whois query against (i.e. example.com)
    -l)       file containing a list of domains to send whois queries to (one domain per line)
    -o)       file to write results into (JSON format)
    -c)       use colors on output (recommended)
    -q)       don't print banner
    -v)       print more information
    -h)       print help panel
    --debug)  print raw query output (use this if provided domain is valid and it throws an error)

Examples:
    gr-whois -d example.com -v
    gr-whois -l domains.txt -c
    cat domains.txt | gr-whois
    `)
}

func info(color bool, text string) {
	if color {
		fmt.Println("["+magenta("*")+"]", text)
	} else {
		fmt.Println("[*]", text)
	}
}

func printWhois(title string, result *wp.Contact, verbose bool, use_color bool) {
	if result != nil {
		if result.Name != "" {
			if use_color {
				info(use_color, title+": "+green(result.Name))
			} else {
				info(use_color, title+": "+result.Name)
			}

		} else {
			info(use_color, title+": ")
		}

		if result.Organization != "" {
			fmt.Println("\tOrganization: " + result.Organization)
		}
		if result.Street != "" {
			fmt.Println("\tStreet: " + result.Street)
		}
		if result.City != "" {
			fmt.Println("\tCity: " + result.City)
		}
		if result.Province != "" {
			fmt.Println("\tProvince: " + result.Province)
		}
		if result.Country != "" {
			if use_color {
				fmt.Println("\tCountry: " + cyan(result.Country))
			} else {
				fmt.Println("\tCountry: " + result.Country)
			}
		}
		if result.PostalCode != "" {
			fmt.Println("\tPostal Code: " + result.PostalCode)
		}
		if result.Phone != "" {
			if use_color {
				fmt.Println("\tPhone: " + cyan(result.Phone))
			} else {
				fmt.Println("\tPhone: " + result.Phone)
			}
		}
		if result.Fax != "" {
			fmt.Println("\tFax: " + result.Fax)
		}
		if verbose {
			if result.ReferralURL != "" {
				if use_color {
					fmt.Println("\tReferral URL: " + cyan(result.ReferralURL))
				} else {
					fmt.Println("\tReferral URL: " + result.ReferralURL)
				}
			}
			if result.ID != "" {
				fmt.Println("\tID: " + result.ID)
			}
		}

		fmt.Println()
	}
}

func main() {
	var domain string
	var list string
	var output string
	var stdin bool
	var quiet bool
	var verbose bool
	var use_color bool
	var debug bool
	var help bool

	flag.StringVar(&domain, "d", "", "domain to send whois query against (i.e. example.com)")
	flag.StringVar(&list, "l", "", "file containing a list of domains to send whois queries to (one per line)")
	flag.StringVar(&output, "o", "", "output file to write results into (json format)")
	flag.BoolVar(&quiet, "q", false, "don't print banner")
	flag.BoolVar(&verbose, "v", false, "print more information")
	flag.BoolVar(&use_color, "c", false, "use colors on output (recommended)")
	flag.BoolVar(&debug, "debug", false, "print raw query output (use this if provided domain is valid and it throws an error)")
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
	}

	if domain != "" {
		raw, err := whois.Whois(domain)
		if err != nil {
			log.Fatal(err)
		}

		if debug {
			fmt.Println(raw)
		}

		result, err := wp.Parse(raw)
		if err != nil {
			log.Fatal(err)
		}

		// Domain
		if use_color {
			info(use_color, "Domain Name: "+green(result.Domain.Name))
			fmt.Println("\tExtension: " + result.Domain.Extension)
			fmt.Println("\tWhois Server: " + cyan(result.Domain.WhoisServer))
			fmt.Println("\tID: " + cyan(result.Domain.ID))
		} else {
			info(use_color, "Domain Name: "+result.Domain.Name)
			fmt.Println("\tExtension: " + result.Domain.Extension)
			fmt.Println("\tWhois Server: " + result.Domain.WhoisServer)
			fmt.Println("\tID: " + result.Domain.ID)
		}

		fmt.Println("\tStatus:")
		for _, i := range result.Domain.Status {
			fmt.Println("\t  " + i)
		}

		fmt.Println("\tName Servers:")
		for _, i := range result.Domain.NameServers {
			if use_color {
				fmt.Println("\t  " + cyan(i))
			} else {
				fmt.Println("\t  " + i)
			}
		}

		if result.Domain.CreatedDate != "" {
			fmt.Println("\tCreation Date: " + result.Domain.CreatedDate)
		}
		if result.Domain.ExpirationDate != "" {
			fmt.Println("\tExpiration Date: " + result.Domain.ExpirationDate)
		}
		if result.Domain.UpdatedDate != "" {
			fmt.Println("\tUpdate Date: " + result.Domain.UpdatedDate)
		}

		if verbose {
			if result.Domain.Punycode != "" {
				if use_color {
					fmt.Println("\tPunycode: " + cyan(result.Domain.Punycode))
				} else {
					fmt.Println("\tPunycode: " + result.Domain.Punycode)
				}
			}
			if result.Domain.DNSSec {
				fmt.Println("\tDNSSec:", result.Domain.DNSSec)
			}
		}

		fmt.Println()
		printWhois("Registrar", result.Registrar, verbose, use_color)
		printWhois("Registrant", result.Registrant, verbose, use_color)
		if verbose {
			printWhois("Administrative", result.Administrative, verbose, use_color)
			printWhois("Technical", result.Technical, verbose, use_color)
		}

		if output != "" {
			json_body, err := json.Marshal(result)
			if err != nil {
				log.Fatal(err)
			}

			_, err = out_f.WriteString(string(json_body))
			if err != nil {
				log.Fatal(err)
			}
		}

	} else if (list != "") || (stdin) {

		var f *os.File
		var err error

		if list != "" {
			f, err = os.Open(list)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

		} else {
			f = os.Stdin
		}

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {

			var result wp.WhoisInfo
			var raw string
			var err error

			// Check if domain format is valid
			if (scanner.Text() != "") && (strings.Contains(scanner.Text(), ".")) {
				if use_color {
					fmt.Println("------- " + magenta(scanner.Text()) + " -------")
				} else {
					fmt.Println("------- " + scanner.Text() + " -------")
				}

				raw, err = whois.Whois(scanner.Text())
				if err != nil {
					log.Fatal(err)
				}

				result, err = wp.Parse(raw)
				if err != nil {
					log.Fatal(err)
				}

			} else {
				if use_color {
					fmt.Println("[" + red("!") + "] Invalid domain format found, skipping line\n")
				} else {
					fmt.Println("[!] Invalid domain format found, skipping line\n")
				}
				continue
			}

			// Domain
			if use_color {
				info(use_color, "Domain Name: "+green(result.Domain.Name))
				fmt.Println("\tExtension: " + result.Domain.Extension)
				fmt.Println("\tWhois Server: " + cyan(result.Domain.WhoisServer))
				fmt.Println("\tID: " + cyan(result.Domain.ID))
			} else {
				info(use_color, "Domain Name: "+result.Domain.Name)
				fmt.Println("\tExtension: " + result.Domain.Extension)
				fmt.Println("\tWhois Server: " + result.Domain.WhoisServer)
				fmt.Println("\tID: " + result.Domain.ID)
			}

			fmt.Println("\tStatus:")
			for _, i := range result.Domain.Status {
				fmt.Println("\t  " + i)
			}

			fmt.Println("\tName Servers:")
			for _, i := range result.Domain.NameServers {
				if use_color {
					fmt.Println("\t  " + cyan(i))
				} else {
					fmt.Println("\t  " + i)
				}
			}

			if result.Domain.CreatedDate != "" {
				fmt.Println("\tCreation Date: " + result.Domain.CreatedDate)
			}
			if result.Domain.ExpirationDate != "" {
				fmt.Println("\tExpiration Date: " + result.Domain.ExpirationDate)
			}
			if result.Domain.UpdatedDate != "" {
				fmt.Println("\tUpdate Date: " + result.Domain.UpdatedDate)
			}

			if verbose {
				if result.Domain.Punycode != "" {
					if use_color {
						fmt.Println("\tPunycode: " + cyan(result.Domain.Punycode))
					} else {
						fmt.Println("\tPunycode: " + result.Domain.Punycode)
					}
				}
				if result.Domain.DNSSec {
					fmt.Println("\tDNSSec:", result.Domain.DNSSec)
				}
			}

			fmt.Println()
			printWhois("Registrar", result.Registrar, verbose, use_color)
			printWhois("Registrant", result.Registrant, verbose, use_color)
			if verbose {
				printWhois("Administrative", result.Administrative, verbose, use_color)
				printWhois("Technical", result.Technical, verbose, use_color)
			}
		}
	}

	if output != "" {
		core.Green("Info written to "+output, use_color)
	}

	if use_color {
		fmt.Println("["+green("+")+"] Elapsed time:", green(core.TimerDiff(t1)))
	} else {
		fmt.Println("[+] Elapsed time:", core.TimerDiff(t1))
	}
}
