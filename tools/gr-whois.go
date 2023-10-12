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
  INPUT:
    -d, -domain string      domain to send whois query against (i.e. example.com)
    -l, -list string        file containing a list of domains to send whois queries to (one domain per line)

  OUTPUT:
    -o, -output string      file to write results into (JSON format)

  CONFIG:
    -v, -verbose      print more information
    -c, -color        use colors on output (recommended)
    -q, -quiet        don't print banner

  DEBUG:
    -debug        print raw query output (use this if provided domain is valid and it throws an error)
    -version      show go-recon version
    -h, -help     print help panel

Examples:
    gr-whois -d example.com -v -o whois.json
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

// nolint: gocyclo
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

// nolint: gocyclo
func main() {
	var domain string
	var list string
	var json_output string
	var quiet bool
	var verbose bool
	var use_color bool
	var debug bool
	var version bool
	var help bool
	var stdin bool

	flag.StringVar(&domain, "d", "", "")
	flag.StringVar(&domain, "domain", "", "")
	flag.StringVar(&list, "l", "", "")
	flag.StringVar(&list, "list", "", "")
	flag.StringVar(&json_output, "o", "", "")
	flag.StringVar(&json_output, "output", "", "")
	flag.BoolVar(&verbose, "v", false, "")
	flag.BoolVar(&verbose, "verbose", false, "")
	flag.BoolVar(&quiet, "q", false, "")
	flag.BoolVar(&quiet, "quiet", false, "")
	flag.BoolVar(&use_color, "c", false, "")
	flag.BoolVar(&use_color, "color", false, "")
	flag.BoolVar(&debug, "debug", false, "")
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

	if (domain == "") && (list == "") && (!stdin) {
		helpPanel()
		os.Exit(0)
	}

	if (domain != "") && (list != "") {
		helpPanel()
		core.Red("You can't use (-d) and (-l) at same time", use_color)
		os.Exit(0)
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

		if json_output != "" {
			json_body, err := json.Marshal(result)
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
				core.Red("Invalid domain format found, skipping line\n", use_color)
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

	if json_output != "" {
		core.Green("Info written to "+json_output+" (JSON)", use_color)
	}

	if use_color {
		fmt.Println("["+green("+")+"] Elapsed time:", green(core.TimerDiff(t1)))
	} else {
		fmt.Println("[+] Elapsed time:", core.TimerDiff(t1))
	}
}
