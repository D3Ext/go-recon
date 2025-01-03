package main

import (
	"bufio"
  "strings"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"sync"

	"github.com/D3Ext/go-recon/core"
)

var red func(a ...interface{}) string = color.New(color.FgRed).SprintFunc()
var cyan func(a ...interface{}) string = color.New(color.FgCyan).SprintFunc()
var green func(a ...interface{}) string = color.New(color.FgGreen).SprintFunc()
var magenta func(a ...interface{}) string = color.New(color.FgMagenta).SprintFunc()
var yellow func(a ...interface{}) string = color.New(color.FgYellow).SprintFunc()

type SecretsInfo struct {
	Secrets []string `json:"secrets"`
	Length  int      `json:"length"`
	Time    string   `json:"time"`
}

func helpPanel() {
	fmt.Println(`Usage of gr-secrets:
  INPUT:
    -u, -url string       url in which to look for secrets (i.e. https://example.com/script.js)
    -l, -list string      file containing a list of JS endpoints to search for secrets (one url per line)

  REGEX:
    -r, -regex string           custom regex to search for (i.e. apikey=secret[a-z]+)
    -rl, -regex-list string     file containing a list of custom regex to search for (one regex per line)

  OUTPUT:
    -o, -output string            file to write secrets into
    -oj, -output-json string      file to write secrets into (JSON format)
    -oc, -output-csv string       file to write secrets into (CSV format)

  CONFIG:
    -w, -workers int      number of concurrent workers (default=10)
    -H, -header string    include custom headers (separated by semicolon) on HTTP requests
    -a, -agent string     user agent to include on requests (default=generic agent)
    -p, -proxy string     proxy to send requests through (i.e. http://127.0.0.1:8080)
    -t, -timeout int      milliseconds to wait before each request timeout (default=5000)
    -c, -color            print colors on output
    -q, -quiet            print neither banner nor logging, only print output

  DEBUG:
    -version      show go-recon version
    -h, -help     print help panel

Examples:
    gr-secrets -u https://example.com -o secrets.txt -c
    gr-secrets -l urls.txt -w 15 -t 6000
    gr-secrets -u https://example.com -rl regexs.txt
    cat js.txt | gr-secrets -r "secret=api_[A-Z]+"
    `)
}

var csv_info [][]string

// nolint: gocyclo
func main() {
	var url string
	var list string
	var regex string
	var regex_list string
	var workers int
  var header string
  var user_agent string
	var proxy string
	var output string
	var json_output string
	var csv_output string
	var timeout int
	var quiet bool
	var use_color bool
	var version bool
	var help bool
	var stdin bool

	flag.StringVar(&url, "u", "", "")
	flag.StringVar(&url, "url", "", "")
	flag.StringVar(&list, "l", "", "")
	flag.StringVar(&list, "list", "", "")
	flag.StringVar(&regex, "r", "", "")
	flag.StringVar(&regex, "regex", "", "")
	flag.StringVar(&regex_list, "rl", "", "")
	flag.StringVar(&regex_list, "regex-list", "", "")
	flag.IntVar(&workers, "w", 10, "")
	flag.IntVar(&workers, "workers", 10, "")
  flag.StringVar(&header, "H", "", "")
  flag.StringVar(&header, "header", "", "")
  flag.StringVar(&user_agent, "a", "Mozilla/5.0 (X11; Linux x86_64; rv:78.0) Gecko/20100101 Firefox/78.0", "")
  flag.StringVar(&user_agent, "agent", "Mozilla/5.0 (X11; Linux x86_64; rv:78.0) Gecko/20100101 Firefox/78.0", "")
	flag.StringVar(&proxy, "p", "", "")
	flag.StringVar(&proxy, "proxy", "", "")
	flag.StringVar(&output, "o", "", "")
	flag.StringVar(&output, "output", "", "")
	flag.StringVar(&json_output, "oj", "", "")
	flag.StringVar(&json_output, "output-json", "", "")
	flag.StringVar(&csv_output, "oc", "", "")
	flag.StringVar(&csv_output, "output-csv", "", "")
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

	if (url == "") && (list == "") && (!stdin) { // check if required arguments were given
		helpPanel()
		os.Exit(0)
	}

	if proxy != "" {
		os.Setenv("HTTP_PROXY", proxy)
		os.Setenv("HTTPS_PROXY", proxy)
	}

	// define variables which will be used to write secrets to file
	var out_f *os.File
	if output != "" {
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

	var found_secrets []string
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		if json_output != "" {
			json_secrets := SecretsInfo{
				Secrets: found_secrets,
				Length:  len(found_secrets),
				Time:    core.TimerDiff(t1).String(),
			}

			json_body, err := json.Marshal(json_secrets)
			if err != nil {
				log.Fatal(err)
			}

			_, err = out_f.WriteString(string(json_body))
			if err != nil {
				log.Fatal(err)
			}
		}
		os.Exit(0)
	}()

	var counter int
	var secrets_list = []string{"(xox[p|b|o|a]-[0-9]{12}-[0-9]{12}-[0-9]{12}-[a-z0-9]{32})", "https://hooks.slack.com/services/T[a-zA-Z0-9_]{8}/B[a-zA-Z0-9_]{8}/[a-zA-Z0-9_]{24}", "[f|F][a|A][c|C][e|E][b|B][o|O][o|O][k|K].{0,30}['\"\\s][0-9a-f]{32}['\"\\s]", "[t|T][w|W][i|I][t|T][t|T][e|E][r|R].{0,30}['\"\\s][0-9a-zA-Z]{35,44}['\"\\s]", "[h|H][e|E][r|R][o|O][k|K][u|U].{0,30}[0-9A-F]{8}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{12}", "key-[0-9a-zA-Z]{32}", "[0-9a-f]{32}-us[0-9]{1,2}", "sk_live_[0-9a-z]{32}", "[0-9(+-[0-9A-Za-z_]{32}.apps.qooqleusercontent.com", "AIza[0-9A-Za-z-_]{35}", "6L[0-9A-Za-z-_]{38}", "ya29\\.[0-9A-Za-z\\-_]+", "AKIA[0-9A-Z]{16}", "amzn\\.mws\\.[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}", "s3\\.amazonaws.com[/]+|[a-zA-Z0-9_-]*\\.s3\\.amazonaws.com", "EAACEdEose0cBA[0-9A-Za-z]+", "key-[0-9a-zA-Z]{32}", "SK[0-9a-fA-F]{32}", "AC[a-zA-Z0-9_\\-]{32}", "AP[a-zA-Z0-9_\\-]{32}", "access_token\\$production\\$[0-9a-z]{16}\\$[0-9a-f]{32}", "sq0csp-[ 0-9A-Za-z\\-_]{43}", "sqOatp-[0-9A-Za-z\\-_]{22}", "sk_live_[0-9a-zA-Z]{24}", "rk_live_[0-9a-zA-Z]{24}", "[a-zA-Z0-9_-]*:[a-zA-Z0-9_\\-]+@github\\.com*", "-----BEGIN PRIVATE KEY-----[a-zA-Z0-9\\S]{100,}-----END PRIVATE KEY-----", "-----BEGIN RSA PRIVATE KEY-----[a-zA-Z0-9\\S]{100,}-----END RSA PRIVATE KEY-----", "(?i)[\"']?zopim[_-]?account[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?zhuliang[_-]?gh[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?zensonatypepassword[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)zendesk(_api_token|_key|_token|-travis-github|_url|_username)(\\s|=)", "(?i)[\"']?yt[_-]?server[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?yt[_-]?partner[_-]?refresh[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?yt[_-]?partner[_-]?client[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?yt[_-]?client[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?yt[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?yt[_-]?account[_-]?refresh[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?yt[_-]?account[_-]?client[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?yangshun[_-]?gh[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?yangshun[_-]?gh[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?www[_-]?googleapis[_-]?com[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?wpt[_-]?ssh[_-]?private[_-]?key[_-]?base64[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?wpt[_-]?ssh[_-]?connect[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?wpt[_-]?report[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?wpt[_-]?prepare[_-]?dir[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?wpt[_-]?db[_-]?user[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?wpt[_-]?db[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?wporg[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?wpjm[_-]?phpunit[_-]?google[_-]?geocode[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?wordpress[_-]?db[_-]?user[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?wordpress[_-]?db[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?wincert[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?widget[_-]?test[_-]?server[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?widget[_-]?fb[_-]?password[_-]?3[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?widget[_-]?fb[_-]?password[_-]?2[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?widget[_-]?fb[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?widget[_-]?basic[_-]?password[_-]?5[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?widget[_-]?basic[_-]?password[_-]?4[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?widget[_-]?basic[_-]?password[_-]?3[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?widget[_-]?basic[_-]?password[_-]?2[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?widget[_-]?basic[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?watson[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?watson[_-]?device[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?watson[_-]?conversation[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?wakatime[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?vscetoken[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?visual[_-]?recognition[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?virustotal[_-]?apikey[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?vip[_-]?github[_-]?deploy[_-]?key[_-]?pass[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?vip[_-]?github[_-]?deploy[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?vip[_-]?github[_-]?build[_-]?repo[_-]?deploy[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?v[_-]?sfdc[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?v[_-]?sfdc[_-]?client[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?usertravis[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?user[_-]?assets[_-]?secret[_-]?access[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?user[_-]?assets[_-]?access[_-]?key[_-]?id[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?use[_-]?ssh[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?us[_-]?east[_-]?1[_-]?elb[_-]?amazonaws[_-]?com[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?urban[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?urban[_-]?master[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?urban[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?unity[_-]?serial[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?unity[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?twitteroauthaccesstoken[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?twitteroauthaccesssecret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?twitter[_-]?consumer[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?twitter[_-]?consumer[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?twine[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?twilio[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?twilio[_-]?sid[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?twilio[_-]?configuration[_-]?sid[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?twilio[_-]?chat[_-]?account[_-]?api[_-]?service[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?twilio[_-]?api[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?twilio[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?trex[_-]?okta[_-]?client[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?trex[_-]?client[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?travis[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?travis[_-]?secure[_-]?env[_-]?vars[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?travis[_-]?pull[_-]?request[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?travis[_-]?gh[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?travis[_-]?e2e[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?travis[_-]?com[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?travis[_-]?branch[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?travis[_-]?api[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?travis[_-]?access[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?token[_-]?core[_-]?java[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?thera[_-]?oss[_-]?access[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?tester[_-]?keys[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?test[_-]?test[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?test[_-]?github[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?tesco[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?svn[_-]?pass[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?surge[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?surge[_-]?login[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?stripe[_-]?public[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?stripe[_-]?private[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?strip[_-]?secret[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?strip[_-]?publishable[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?stormpath[_-]?api[_-]?key[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?stormpath[_-]?api[_-]?key[_-]?id[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?starship[_-]?auth[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?starship[_-]?account[_-]?sid[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?star[_-]?test[_-]?secret[_-]?access[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?star[_-]?test[_-]?location[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?star[_-]?test[_-]?bucket[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?star[_-]?test[_-]?aws[_-]?access[_-]?key[_-]?id[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?staging[_-]?base[_-]?url[_-]?runscope[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?ssmtp[_-]?config[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?sshpass[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?srcclr[_-]?api[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?square[_-]?reader[_-]?sdk[_-]?repository[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?sqssecretkey[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?sqsaccesskey[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?spring[_-]?mail[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?spotify[_-]?api[_-]?client[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?spotify[_-]?api[_-]?access[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?spaces[_-]?secret[_-]?access[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?spaces[_-]?access[_-]?key[_-]?id[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?soundcloud[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?soundcloud[_-]?client[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?sonatypepassword[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?sonatype[_-]?token[_-]?user[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?sonatype[_-]?token[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?sonatype[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?sonatype[_-]?pass[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?sonatype[_-]?nexus[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?sonatype[_-]?gpg[_-]?passphrase[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?sonatype[_-]?gpg[_-]?key[_-]?name[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?sonar[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?sonar[_-]?project[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?sonar[_-]?organization[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?socrata[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?socrata[_-]?app[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?snyk[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?snyk[_-]?api[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?snoowrap[_-]?refresh[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?snoowrap[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?snoowrap[_-]?client[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?slate[_-]?user[_-]?email[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?slash[_-]?developer[_-]?space[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?slash[_-]?developer[_-]?space[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?signing[_-]?key[_-]?sid[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?signing[_-]?key[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?signing[_-]?key[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?signing[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?setsecretkey[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?setdstsecretkey[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?setdstaccesskey[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?ses[_-]?secret[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?ses[_-]?access[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?service[_-]?account[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?sentry[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?sentry[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?sentry[_-]?endpoint[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?sentry[_-]?default[_-]?org[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?sentry[_-]?auth[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?sendwithus[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?sendgrid[_-]?username[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?sendgrid[_-]?user[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?sendgrid[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?sendgrid[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?sendgrid[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?sendgrid[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?selion[_-]?selenium[_-]?host[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?selion[_-]?log[_-]?level[_-]?dev[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?segment[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?secretkey[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?secretaccesskey[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?secret[_-]?key[_-]?base[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?secret[_-]?9[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?secret[_-]?8[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?secret[_-]?7[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?secret[_-]?6[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?secret[_-]?5[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?secret[_-]?4[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?secret[_-]?3[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?secret[_-]?2[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?secret[_-]?11[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?secret[_-]?10[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?secret[_-]?1[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?secret[_-]?0[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?sdr[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?scrutinizer[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?sauce[_-]?access[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?sandbox[_-]?aws[_-]?secret[_-]?access[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?sandbox[_-]?aws[_-]?access[_-]?key[_-]?id[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?sandbox[_-]?access[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?salesforce[_-]?bulk[_-]?test[_-]?security[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?salesforce[_-]?bulk[_-]?test[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?sacloud[_-]?api[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?sacloud[_-]?access[_-]?token[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?sacloud[_-]?access[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?s3[_-]?user[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?s3[_-]?secret[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?s3[_-]?secret[_-]?assets[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?s3[_-]?secret[_-]?app[_-]?logs[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?s3[_-]?key[_-]?assets[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?s3[_-]?key[_-]?app[_-]?logs[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?s3[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?s3[_-]?external[_-]?3[_-]?amazonaws[_-]?com[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?s3[_-]?bucket[_-]?name[_-]?assets[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?s3[_-]?bucket[_-]?name[_-]?app[_-]?logs[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?s3[_-]?access[_-]?key[_-]?id[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?s3[_-]?access[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?rubygems[_-]?auth[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?rtd[_-]?store[_-]?pass[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?rtd[_-]?key[_-]?pass[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?route53[_-]?access[_-]?key[_-]?id[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?ropsten[_-]?private[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?rinkeby[_-]?private[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?rest[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?repotoken[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?reporting[_-]?webdav[_-]?url[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?reporting[_-]?webdav[_-]?pwd[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?release[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?release[_-]?gh[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?registry[_-]?secure[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?registry[_-]?pass[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?refresh[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?rediscloud[_-]?url[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?redis[_-]?stunnel[_-]?urls[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?randrmusicapiaccesstoken[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?rabbitmq[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?quip[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?qiita[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?pypi[_-]?passowrd[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?pushover[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?publish[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?publish[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?publish[_-]?access[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?project[_-]?config[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?prod[_-]?secret[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?prod[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?prod[_-]?access[_-]?key[_-]?id[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?private[_-]?signing[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?pring[_-]?mail[_-]?username[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?preferred[_-]?username[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?prebuild[_-]?auth[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?postgresql[_-]?pass[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?postgresql[_-]?db[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?postgres[_-]?env[_-]?postgres[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?postgres[_-]?env[_-]?postgres[_-]?db[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?plugin[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?plotly[_-]?apikey[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?places[_-]?apikey[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?places[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?pg[_-]?host[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?pg[_-]?database[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?personal[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?personal[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?percy[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?percy[_-]?project[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?paypal[_-]?client[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?passwordtravis[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?parse[_-]?js[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?pagerduty[_-]?apikey[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?packagecloud[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?ossrh[_-]?username[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?ossrh[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?ossrh[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?ossrh[_-]?pass[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?ossrh[_-]?jira[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?os[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?os[_-]?auth[_-]?url[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?org[_-]?project[_-]?gradle[_-]?sonatype[_-]?nexus[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?org[_-]?gradle[_-]?project[_-]?sonatype[_-]?nexus[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?openwhisk[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?open[_-]?whisk[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?onesignal[_-]?user[_-]?auth[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?onesignal[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?omise[_-]?skey[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?omise[_-]?pubkey[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?omise[_-]?pkey[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?omise[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?okta[_-]?oauth2[_-]?clientsecret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?okta[_-]?oauth2[_-]?client[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?okta[_-]?client[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?ofta[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?ofta[_-]?region[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?ofta[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?octest[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?octest[_-]?app[_-]?username[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?octest[_-]?app[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?oc[_-]?pass[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?object[_-]?store[_-]?creds[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?object[_-]?store[_-]?bucket[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?object[_-]?storage[_-]?region[_-]?name[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?object[_-]?storage[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?oauth[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?numbers[_-]?service[_-]?pass[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?nuget[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?nuget[_-]?apikey[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?nuget[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?npm[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?npm[_-]?secret[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?npm[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?npm[_-]?email[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?npm[_-]?auth[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?npm[_-]?api[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?npm[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?now[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?non[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?node[_-]?pre[_-]?gyp[_-]?secretaccesskey[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?node[_-]?pre[_-]?gyp[_-]?github[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?node[_-]?pre[_-]?gyp[_-]?accesskeyid[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?node[_-]?env[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?ngrok[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?ngrok[_-]?auth[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?nexuspassword[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?nexus[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?new[_-]?relic[_-]?beta[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?netlify[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?nativeevents[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?mysqlsecret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?mysqlmasteruser[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?mysql[_-]?username[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?mysql[_-]?user[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?mysql[_-]?root[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?mysql[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?mysql[_-]?hostname[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?mysql[_-]?database[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?my[_-]?secret[_-]?env[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?multi[_-]?workspace[_-]?sid[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?multi[_-]?workflow[_-]?sid[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?multi[_-]?disconnect[_-]?sid[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?multi[_-]?connect[_-]?sid[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?multi[_-]?bob[_-]?sid[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?minio[_-]?secret[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?minio[_-]?access[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?mile[_-]?zero[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?mh[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?mh[_-]?apikey[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?mg[_-]?public[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?mg[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?mapboxaccesstoken[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?mapbox[_-]?aws[_-]?secret[_-]?access[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?mapbox[_-]?aws[_-]?access[_-]?key[_-]?id[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?mapbox[_-]?api[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?mapbox[_-]?access[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?manifest[_-]?app[_-]?url[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?manifest[_-]?app[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?mandrill[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?managementapiaccesstoken[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?management[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?manage[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?manage[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?mailgun[_-]?secret[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?mailgun[_-]?pub[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?mailgun[_-]?pub[_-]?apikey[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?mailgun[_-]?priv[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?mailgun[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?mailgun[_-]?apikey[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?mailgun[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?mailer[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?mailchimp[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?mailchimp[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?mail[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?magento[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?magento[_-]?auth[_-]?username [\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?magento[_-]?auth[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?lottie[_-]?upload[_-]?cert[_-]?key[_-]?store[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?lottie[_-]?upload[_-]?cert[_-]?key[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?lottie[_-]?s3[_-]?secret[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?lottie[_-]?happo[_-]?secret[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?lottie[_-]?happo[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?looker[_-]?test[_-]?runner[_-]?client[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?ll[_-]?shared[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?ll[_-]?publish[_-]?url[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?linux[_-]?signing[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?linkedin[_-]?client[_-]?secretor lottie[_-]?s3[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?lighthouse[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?lektor[_-]?deploy[_-]?username[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?lektor[_-]?deploy[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?leanplum[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?kxoltsn3vogdop92m[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?kubeconfig[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?kubecfg[_-]?s3[_-]?path[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?kovan[_-]?private[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?keystore[_-]?pass[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?kafka[_-]?rest[_-]?url[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?kafka[_-]?instance[_-]?name[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?kafka[_-]?admin[_-]?url[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?jwt[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?jdbc:mysql[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?jdbc[_-]?host[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?jdbc[_-]?databaseurl[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?itest[_-]?gh[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?ios[_-]?docs[_-]?deploy[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?internal[_-]?secrets[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?integration[_-]?test[_-]?appid[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?integration[_-]?test[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?index[_-]?name[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?ij[_-]?repo[_-]?username[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?ij[_-]?repo[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?hub[_-]?dxia2[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?homebrew[_-]?github[_-]?api[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?hockeyapp[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?heroku[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?heroku[_-]?email[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?heroku[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?hb[_-]?codesign[_-]?key[_-]?pass[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?hb[_-]?codesign[_-]?gpg[_-]?pass[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?hab[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?hab[_-]?auth[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?grgit[_-]?user[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?gren[_-]?github[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?gradle[_-]?signing[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?gradle[_-]?signing[_-]?key[_-]?id[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?gradle[_-]?publish[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?gradle[_-]?publish[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?gpg[_-]?secret[_-]?keys[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?gpg[_-]?private[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?gpg[_-]?passphrase[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?gpg[_-]?ownertrust[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?gpg[_-]?keyname[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?gpg[_-]?key[_-]?name[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?google[_-]?private[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?google[_-]?maps[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?google[_-]?client[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?google[_-]?client[_-]?id[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?google[_-]?client[_-]?email[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?google[_-]?account[_-]?type[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?gogs[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?gitlab[_-]?user[_-]?email[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?github[_-]?tokens[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?github[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?github[_-]?repo[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?github[_-]?release[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?github[_-]?pwd[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?github[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?github[_-]?oauth[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?github[_-]?oauth[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?github[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?github[_-]?hunter[_-]?username[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?github[_-]?hunter[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?github[_-]?deployment[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?github[_-]?deploy[_-]?hb[_-]?doc[_-]?pass[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?github[_-]?client[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?github[_-]?auth[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?github[_-]?auth[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?github[_-]?api[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?github[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?github[_-]?access[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?git[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?git[_-]?name[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?git[_-]?email[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?git[_-]?committer[_-]?name[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?git[_-]?committer[_-]?email[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?git[_-]?author[_-]?name[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?git[_-]?author[_-]?email[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?ghost[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?ghb[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?gh[_-]?unstable[_-]?oauth[_-]?client[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?gh[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?gh[_-]?repo[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?gh[_-]?oauth[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?gh[_-]?oauth[_-]?client[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?gh[_-]?next[_-]?unstable[_-]?oauth[_-]?client[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?gh[_-]?next[_-]?unstable[_-]?oauth[_-]?client[_-]?id[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?gh[_-]?next[_-]?oauth[_-]?client[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?gh[_-]?email[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?gh[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?gcs[_-]?bucket[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?gcr[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?gcloud[_-]?service[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?gcloud[_-]?project[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?gcloud[_-]?bucket[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?ftp[_-]?username[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?ftp[_-]?user[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?ftp[_-]?pw[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?ftp[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?ftp[_-]?login[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?ftp[_-]?host[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?fossa[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?flickr[_-]?api[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?flickr[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?flask[_-]?secret[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?firefox[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?firebase[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?firebase[_-]?project[_-]?develop[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?firebase[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?firebase[_-]?api[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?firebase[_-]?api[_-]?json[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?file[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?exp[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?eureka[_-]?awssecretkey[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?env[_-]?sonatype[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?env[_-]?secret[_-]?access[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?env[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?env[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?env[_-]?heroku[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?env[_-]?github[_-]?oauth[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?end[_-]?user[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?encryption[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?elasticsearch[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?elastic[_-]?cloud[_-]?auth[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?dsonar[_-]?projectkey[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?dsonar[_-]?login[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?droplet[_-]?travis[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?dropbox[_-]?oauth[_-]?bearer[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?doordash[_-]?auth[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?dockerhubpassword[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?dockerhub[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?docker[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?docker[_-]?postgres[_-]?url[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?docker[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?docker[_-]?passwd[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?docker[_-]?pass[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?docker[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?docker[_-]?hub[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?digitalocean[_-]?ssh[_-]?key[_-]?ids[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?digitalocean[_-]?ssh[_-]?key[_-]?body[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?digitalocean[_-]?access[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?dgpg[_-]?passphrase[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?deploy[_-]?user[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?deploy[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?deploy[_-]?secure[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?deploy[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?ddgc[_-]?github[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?ddg[_-]?test[_-]?email[_-]?pw[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?ddg[_-]?test[_-]?email[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?db[_-]?username[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?db[_-]?user[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?db[_-]?pw[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?db[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?db[_-]?host[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?db[_-]?database[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?db[_-]?connection[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?datadog[_-]?app[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?datadog[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?database[_-]?username[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?database[_-]?user[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?database[_-]?port[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?database[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?database[_-]?name[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?database[_-]?host[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?danger[_-]?github[_-]?api[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?cypress[_-]?record[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?coverity[_-]?scan[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?coveralls[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?coveralls[_-]?repo[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?coveralls[_-]?api[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?cos[_-]?secrets[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?conversation[_-]?username[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?conversation[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?contentful[_-]?v2[_-]?access[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?contentful[_-]?test[_-]?org[_-]?cma[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?contentful[_-]?php[_-]?management[_-]?test[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?contentful[_-]?management[_-]?api[_-]?access[_-]?token[_-]?new[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?contentful[_-]?management[_-]?api[_-]?access[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?contentful[_-]?integration[_-]?management[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?contentful[_-]?cma[_-]?test[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?contentful[_-]?access[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?consumerkey[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?consumer[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?conekta[_-]?apikey[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?coding[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?codecov[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?codeclimate[_-]?repo[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?codacy[_-]?project[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?cocoapods[_-]?trunk[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?cocoapods[_-]?trunk[_-]?email[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?cn[_-]?secret[_-]?access[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?cn[_-]?access[_-]?key[_-]?id[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?clu[_-]?ssh[_-]?private[_-]?key[_-]?base64[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?clu[_-]?repo[_-]?url[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?cloudinary[_-]?url[_-]?staging[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?cloudinary[_-]?url[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?cloudflare[_-]?email[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?cloudflare[_-]?auth[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?cloudflare[_-]?auth[_-]?email[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?cloudflare[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?cloudant[_-]?service[_-]?database[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?cloudant[_-]?processed[_-]?database[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?cloudant[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?cloudant[_-]?parsed[_-]?database[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?cloudant[_-]?order[_-]?database[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?cloudant[_-]?instance[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?cloudant[_-]?database[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?cloudant[_-]?audited[_-]?database[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?cloudant[_-]?archived[_-]?database[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?cloud[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?clojars[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?client[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?cli[_-]?e2e[_-]?cma[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?claimr[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?claimr[_-]?superuser[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?claimr[_-]?db[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?claimr[_-]?database[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?ci[_-]?user[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?ci[_-]?server[_-]?name[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?ci[_-]?registry[_-]?user[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?ci[_-]?project[_-]?url[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?ci[_-]?deploy[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?chrome[_-]?refresh[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?chrome[_-]?client[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?cheverny[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?cf[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?certificate[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?censys[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?cattle[_-]?secret[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?cattle[_-]?agent[_-]?instance[_-]?auth[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?cattle[_-]?access[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?cargo[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?cache[_-]?s3[_-]?secret[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?bx[_-]?username[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?bx[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?bundlesize[_-]?github[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?built[_-]?branch[_-]?deploy[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?bucketeer[_-]?aws[_-]?secret[_-]?access[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?bucketeer[_-]?aws[_-]?access[_-]?key[_-]?id[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?browserstack[_-]?access[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?browser[_-]?stack[_-]?access[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?brackets[_-]?repo[_-]?oauth[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?bluemix[_-]?username[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?bluemix[_-]?pwd[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?bluemix[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?bluemix[_-]?pass[_-]?prod[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?bluemix[_-]?pass[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?bluemix[_-]?auth[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?bluemix[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?bintraykey[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?bintray[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?bintray[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?bintray[_-]?gpg[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?bintray[_-]?apikey[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?bintray[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?b2[_-]?bucket[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?b2[_-]?app[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?awssecretkey[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?awscn[_-]?secret[_-]?access[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?awscn[_-]?access[_-]?key[_-]?id[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?awsaccesskeyid[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?aws[_-]?ses[_-]?secret[_-]?access[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?aws[_-]?ses[_-]?access[_-]?key[_-]?id[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?aws[_-]?secrets[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?aws[_-]?secret[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?aws[_-]?secret[_-]?access[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?aws[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?aws[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?aws[_-]?config[_-]?secretaccesskey[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?aws[_-]?config[_-]?accesskeyid[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?aws[_-]?access[_-]?key[_-]?id[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?aws[_-]?access[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?aws[_-]?access[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?author[_-]?npm[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?author[_-]?email[_-]?addr[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?auth0[_-]?client[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?auth0[_-]?api[_-]?clientsecret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?auth[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?assistant[_-]?iam[_-]?apikey[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?artifacts[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?artifacts[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?artifacts[_-]?bucket[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?artifacts[_-]?aws[_-]?secret[_-]?access[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?artifacts[_-]?aws[_-]?access[_-]?key[_-]?id[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?artifactory[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?argos[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?apple[_-]?id[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?appclientsecret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?app[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?app[_-]?secrete[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?app[_-]?report[_-]?token[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?app[_-]?bucket[_-]?perm[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?apigw[_-]?access[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?apiary[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?api[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?api[_-]?key[_-]?sid[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?api[_-]?key[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?aos[_-]?sec[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?aos[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?ansible[_-]?vault[_-]?password[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?android[_-]?docs[_-]?deploy[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?anaconda[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?amazon[_-]?secret[_-]?access[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?amazon[_-]?bucket[_-]?name[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?alicloud[_-]?secret[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?alicloud[_-]?access[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?alias[_-]?pass[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?algolia[_-]?search[_-]?key[_-]?1[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?algolia[_-]?search[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?algolia[_-]?search[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?algolia[_-]?api[_-]?key[_-]?search[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?algolia[_-]?api[_-]?key[_-]?mcm[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?algolia[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?algolia[_-]?admin[_-]?key[_-]?mcm[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?algolia[_-]?admin[_-]?key[_-]?2[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?algolia[_-]?admin[_-]?key[_-]?1[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?air[-_]?table[-_]?api[-_]?key[\"']?[=:][\"']?.+[\"']", "(?i)[\"']?adzerk[_-]?api[_-]?key[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?admin[_-]?email[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?account[_-]?sid[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?access[_-]?token[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?access[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?", "(?i)[\"']?access[_-]?key[_-]?secret[\"']?[^\\S\r\n]*[=:][^\\S\r\n]*[\"']?[\\w-]+[\"']?"}

	if url != "" {

		if !quiet {
			core.Warning("Use with caution.", use_color)
			core.Magenta("Concurrent workers: "+strconv.Itoa(workers), use_color)
			if proxy != "" {
				core.Magenta("Proxy: "+proxy, use_color)
			}
			core.Magenta("Looking for secrets...", use_color)
		}

		client := core.CreateHttpClient(timeout)
		req, _ := http.NewRequest("GET", url, nil)

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

		resp, err := client.Do(req) // Send request
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		if regex != "" { // enter here if custom regex was provided
			if !quiet {
				core.Magenta("Using custom regex to filter for secrets: "+regex+"\n", use_color)
			}

			match, err := regexp.MatchString(regex, string(body))
			if err != nil {
				log.Fatal(err)
			}

			if match {
				found_pattern := regexp.MustCompile(regex)
				str_match := found_pattern.FindString(string(body))
				found_secrets = append(found_secrets, str_match)
				csv_info = append(csv_info, []string{url, str_match})
				counter += 1

				writeOutput(str_match, output, out_f)

				logSecret(str_match, use_color, quiet)
			}

		} else if regex_list != "" {
			if !quiet {
				core.Magenta("Using custom list of regular expressions...\n", use_color)
			}

			f, err := os.Open(regex_list)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				r := scanner.Text()

				match, err := regexp.MatchString(r, string(body))
				if err != nil {
					log.Fatal(err)
				}

				if match {
					found_pattern := regexp.MustCompile(r)
					str_match := found_pattern.FindString(string(body))
					found_secrets = append(found_secrets, str_match)
					csv_info = append(csv_info, []string{url, str_match})
					counter += 1

					writeOutput(str_match, output, out_f)

					logSecret(str_match, use_color, quiet)
				}
			}

		} else { // enter here if no custom regex was provided
			if !quiet {
				core.Magenta("Using default regular expression...\n", use_color)
			}

			for _, secret := range secrets_list {
				match, err := regexp.MatchString(secret, string(body))
				if err != nil {
					log.Fatal(err)
				}

				if match {
					found_pattern := regexp.MustCompile(secret)
					str_match := found_pattern.FindString(string(body))
					found_secrets = append(found_secrets, str_match)
					csv_info = append(csv_info, []string{url, str_match})
					counter += 1

					writeOutput(str_match, output, out_f)

					logSecret(str_match, use_color, quiet)
				}
			}
		}

	} else if (list != "") || (stdin) {

		if !quiet {
			core.Magenta("Looking for secrets on given urls...", use_color)
		}

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
				for js_endpoint := range urls_c {

					client := core.CreateHttpClient(timeout)
					req, _ := http.NewRequest("GET", js_endpoint, nil)

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

					resp, err := client.Do(req) // Send request
					if err != nil {
						log.Fatal(err)
					}
					defer resp.Body.Close()

					body, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						log.Fatal(err)
					}

					if regex != "" {
						match, err := regexp.MatchString(regex, string(body))
						if err != nil {
							log.Fatal(err)
						}

						if match {
							found_pattern := regexp.MustCompile(regex)
							str_match := found_pattern.FindString(string(body))
							found_secrets = append(found_secrets, str_match)
							csv_info = append(csv_info, []string{js_endpoint, str_match})
							counter += 1

							writeOutput(str_match, output, out_f)

							logSecretLocation(str_match, js_endpoint, use_color, quiet)
						}

					} else if regex_list != "" {
						f, err := os.Open(regex_list)
						if err != nil {
							log.Fatal(err)
						}
						defer f.Close()

						scanner := bufio.NewScanner(f)
						for scanner.Scan() {
							r := scanner.Text()

							match, err := regexp.MatchString(r, string(body))
							if err != nil {
								log.Fatal(err)
							}

							if match {
								found_pattern := regexp.MustCompile(r)
								str_match := found_pattern.FindString(string(body))
								found_secrets = append(found_secrets, str_match)
								csv_info = append(csv_info, []string{js_endpoint, str_match})
								counter += 1

								writeOutput(str_match, output, out_f)

								logSecretLocation(str_match, js_endpoint, use_color, quiet)
							}
						}

					} else {
						for _, secret := range secrets_list {
							match, err := regexp.MatchString(secret, string(body))
							if err != nil {
								log.Fatal(err)
							}

							if match {
								found_pattern := regexp.MustCompile(secret)
								str_match := found_pattern.FindString(string(body))
								found_secrets = append(found_secrets, str_match)
								csv_info = append(csv_info, []string{js_endpoint, str_match})
								counter += 1

								writeOutput(str_match, output, out_f)

								logSecretLocation(str_match, js_endpoint, use_color, quiet)
							}
						}
					}
				}

				wg.Done()
			}()
		}

		var urls_counter int
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if scanner.Text() != "" {
				urls_counter += 1
				urls_c <- scanner.Text()
			}
		}

		close(urls_c)
		wg.Wait()
	}

	if json_output != "" {
		json_secrets := SecretsInfo{
			Secrets: found_secrets,
			Length:  len(found_secrets),
			Time:    core.TimerDiff(t1).String(),
		}

		json_body, err := json.Marshal(json_secrets)
		if err != nil {
			log.Fatal(err)
		}

		_, err = out_f.WriteString(string(json_body))
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

		headers := []string{"urls", "secrets"}
		writer.Write(headers)

		for _, row := range csv_info {
			writer.Write(row)
		}
	}

	if !quiet {
		if counter >= 1 {
			fmt.Println()

			if output != "" {
				core.Green("Secrets written to "+output+" (TXT)", use_color)
			}

			if json_output != "" {
				core.Green("Secrets written to "+json_output+" (JSON)", use_color)
			}

			if csv_output != "" {
				core.Green("Secrets written to "+csv_output+" (CSV)", use_color)
			}

		} else {
			fmt.Println()
			core.Red("No secrets found", use_color)
		}

		if use_color {
			fmt.Println("["+green("+")+"]", cyan(counter), "secrets found")
		} else {
			fmt.Println("[+]", counter, "secrets found")
		}

		if use_color {
			fmt.Println("["+green("+")+"] Elapsed time:", green(core.TimerDiff(t1)))
		} else {
			fmt.Println("[+] Elapsed time:", core.TimerDiff(t1))
		}
	}
}

func logSecret(str string, use_color bool, quiet bool) {
	if use_color {
		if !quiet {
			fmt.Println("["+green("+")+"] Secret found:", str)
		} else {
			fmt.Println("["+green("+")+"]", str)
		}
	} else {
		if !quiet {
			fmt.Println("[+] Secret found:", str)
		} else {
			fmt.Println("[+]", str)
		}
	}
}

func logSecretLocation(str string, url string, use_color bool, quiet bool) {
	if use_color {
		if !quiet {
			fmt.Println("["+green("+")+"] Secret found on", url, ":", str)
		} else {
			fmt.Println("["+green("+")+"]", str)
		}
	} else {
		if !quiet {
			fmt.Println("[+] Secret found on", url, ":", str)
		} else {
			fmt.Println("[+]", str)
		}
	}
}

func writeOutput(str string, output string, out_f *os.File) {
	if output != "" {
		_, err := out_f.WriteString(str + "\n")
		if err != nil {
			log.Fatal(err)
		}
	}
}
