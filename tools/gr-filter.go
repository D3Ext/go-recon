package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"log"
	"net/url"
	nurl "net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/D3Ext/go-recon/core"
)

var red func(a ...interface{}) string = color.New(color.FgRed).SprintFunc()
var cyan func(a ...interface{}) string = color.New(color.FgCyan).SprintFunc()
var green func(a ...interface{}) string = color.New(color.FgGreen).SprintFunc()
var magenta func(a ...interface{}) string = color.New(color.FgMagenta).SprintFunc()
var yellow func(a ...interface{}) string = color.New(color.FgYellow).SprintFunc()

type PatternFormat struct {
	Description string   `json:"description,omitempty`
	Flags       string   `json:"flags"`
	Pattern     string   `json:"pattern,omitempty"`
	Patterns    []string `json:"patterns,omitempty"`
}

type FilterInfo struct {
	Urls    []string `json:"urls"`
	Filters []string `json:"filters"`
	Length  int      `json:"length"`
}

func helpPanel() {
	fmt.Println(`Usage of gr-filter:
  INPUT:
    -l, -list string      file containing a list of urls to remove duplicates and useless ones (one domain/url per line)

  OUTPUT:
    -o, -output string          file to write filtered urls/output into (TXT format)
    -oj, -output-json string    file to write filtered urls/output into (JSON format)
    -oc, -output-csv string     file to write filtered urls/output into (CSV format)

  FILTERS:
    -f, -filter string[]    custom filters to use, separated by comma (i.e. nocontent,hasparams)
    -lf, -list-filters      list available patterns/filters

  CONFIG:
    -b, -blacklist string[]   blacklisted extensions to exclude from filtered urls (default extensions are also excluded) (i.e. php,mp3)
    -w, -whitelist string[]   whitelisted extensions to filter for (separated by comma) (i.e. json,txt,php)
    -p, -params string        replace parameter values with given string (i.e. FUZZ)
    -c, -color                print colors on output
    -q, -quiet                print neither banner nor logging, only print output

  DEBUG:
    -version      show go-recon version
    -h, -help     print help panel

Examples:
    gr-filter -l urls.txt -o clean.txt -c
    gr-filter -l urls.txt -b html,asp
    gr-filter -l urls.txt -w zip -oj output.json
    cat urls.txt | gr-filter -f hasparams -o param_urls.txt -q
    cat urls.txt | gr-filter -f redirects,nocontent -p FUZZ
    cat output.txt | gr-filter -f custom_filter
    `)
}

// english words
// nolint: misspell
var text_content = [111]string{"blog", "post", "posts", "stories", "press", "magazine", "news", "articles", "opinions", "images", "comments", "updates", "interviews", "galery", "advices", "story", "stories", "current-affairs", "chronicles", "reports", "reviews", "life", "journal", "travel", "experiencies", "editorial", "publications", "texts", "writings", "tales", "announcements", "analysis", "columns", "topics", "section", "bloggers", "journalism", "notes", "blog-articles"}

var useless_extensions = []string{"png", "jpeg", "gif", "jpg", "svg", "jfif", "avif", "webp", "ico", "tif", "tiff", "ttf", "woff", "mp3", "mp4", "avi", "mov", "mpeg", "wav", "css"}

var blacklist []string

var whitelist []string

var seen_patterns []string

var seen_params []string

var csv_info [][]string

// nolint: gocyclo
func main() {
	var list string
	var blacklist_param string
	var whitelist_param string
	var filter string
	var list_patterns bool
	var params_str string
	var output string
	var json_output string
	var csv_output string
	var quiet bool
	var use_color bool
	var version bool
	var help bool
	var stdin bool

	flag.StringVar(&list, "l", "", "")
	flag.StringVar(&list, "list", "", "")
	flag.StringVar(&blacklist_param, "b", "", "")
	flag.StringVar(&blacklist_param, "blacklist", "", "")
	flag.StringVar(&whitelist_param, "w", "", "")
	flag.StringVar(&whitelist_param, "whitelist", "", "")
	flag.StringVar(&filter, "f", "", "")
	flag.StringVar(&filter, "filter", "", "")
	flag.BoolVar(&list_patterns, "lf", false, "")
	flag.BoolVar(&list_patterns, "list-filters", false, "")
	flag.StringVar(&params_str, "p", "", "")
	flag.StringVar(&params_str, "params", "", "")
	flag.StringVar(&output, "o", "", "")
	flag.StringVar(&output, "output", "", "")
	flag.StringVar(&json_output, "oj", "", "")
	flag.StringVar(&json_output, "output-json", "", "")
	flag.StringVar(&csv_output, "oc", "", "")
	flag.StringVar(&csv_output, "output-csv", "", "")
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

	if list_patterns {
		listPatterns(quiet, use_color)
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

	var counter int
	var custom_filter bool
	var url_based bool
	var outb bytes.Buffer
	filters := strings.Join(strings.Split(strings.TrimSpace(filter), ","), "-")

	if !strings.Contains(filter, "noparam") && !strings.Contains(filter, "hasparam") && !strings.Contains(filter, "noextension") && !strings.Contains(filter, "hasextension") && !strings.Contains(filter, "nocontent") && (filter != "") {

		custom_filter = true
		// get directory where patterns are stored
		patterns_dir, err := getPatternDir()
		if err != nil {
			log.Fatal(err)
		}

		filename := filepath.Join(patterns_dir, filter+".json")
		pattern_f, err := os.Open(filename)
		if err != nil {
			log.Fatal(errors.New("especified pattern doesn't exist under " + patterns_dir))
		}
		defer pattern_f.Close()

		if !quiet {
			core.Magenta("Applying filters on given output...\n", use_color)
		}

		// parse given pattern to filter for
		pattern := PatternFormat{}
		dec := json.NewDecoder(pattern_f)
		err = dec.Decode(&pattern)
		if err != nil {
			log.Fatal(err)
		}

		var patterns_to_filter string
		if len(pattern.Patterns) > 0 {
			patterns_to_filter = "(" + strings.Join(pattern.Patterns, "|") + ")"
		} else if pattern.Pattern != "" {
			patterns_to_filter = pattern.Pattern
		} else {
			log.Fatal(errors.New("empty patterns to filter for"))
		}

		var cmd *exec.Cmd
		operator := "grep"
		if stdin {
			cmd = exec.Command(operator, pattern.Flags, patterns_to_filter)
		} else {
			cmd = exec.Command(operator, pattern.Flags, patterns_to_filter, list)
		}

		cmd.Stdin = os.Stdin
		cmd.Stdout = &outb // save output on buffer for later processing if output is a list of urls
		cmd.Stderr = os.Stderr
		cmd.Run()
	}

	output_slice := strings.Split(outb.String(), "\n")

	if custom_filter {
		if isValidUrl(output_slice[0]) && isValidUrl(output_slice[1]) {
			url_based = true
		} else if len(output_slice) >= 2 && isValidUrl(output_slice[0]) { // if a low number of urls are the results, treat them like a url-based output
			url_based = true
		} else {
			if outb.String() != "" {
				fmt.Println(outb.String())

				if output != "" {
					txt_out, err := os.Create(output)
					if err != nil {
						log.Fatal(err)
					}

					_, err = txt_out.WriteString(outb.String())
					if err != nil {
						log.Fatal(err)
					}
				}
			} else {
				core.Red("No output results returned from filter", use_color)
			}
		}
	}

	// if domain, list and stdin parameters are empty print help panel and exit
	if (list == "") && (!stdin) {
		helpPanel()
		os.Exit(0)
	}

	// check if blacklist and whitelist is provided at same time
	if (blacklist_param != "") && (whitelist_param != "") {
		helpPanel()
		core.Red("You can't use blacklist (-b) and whitelist (-w) at same time", use_color)
		os.Exit(0)
	}

	// define variables which will be used to write results to file
	var txt_out *os.File
	if output != "" {
		txt_out, err = os.Create(output)
		if err != nil {
			log.Fatal(err)
		}
	}

	// before doing anything, check if invalid filter or filters are provided
	if filter != "" {
		if strings.Contains(filter, "hasparam") && strings.Contains(filter, "noparam") {
			core.Red("Invalid filters provided! hasparam and noparam can't be used at the same time", use_color)
			os.Exit(0)

		} else if strings.Contains(filter, "hasextension") && strings.Contains(filter, "noextension") {
			core.Red("Invalid filters provided! hasextension and noextension can't be used at the same time", use_color)
			os.Exit(0)
		}
	}

	if !quiet && !custom_filter {
		core.Magenta("Removing duplicated urls and applying filters...\n", use_color)
	}

	// process blacklist parameter
	if strings.Contains(blacklist_param, ",") { // check if multiple extensions are present
		for _, ext := range strings.Split(blacklist_param, ",") {
			ext = strings.TrimSpace(ext)
			blacklist = append(blacklist, ext)
		}

	} else if blacklist_param != "" { // enter here if there's just one extension
		blacklist = append(blacklist, blacklist_param)
	}

	// process whitelist parameter
	if strings.Contains(whitelist_param, ",") { // check if multiple extensions are present
		for _, ext := range strings.Split(whitelist_param, ",") {
			ext = strings.TrimSpace(ext)
			whitelist = append(whitelist, ext)
		}

	} else if whitelist_param != "" { // enter here if there's just one extension
		whitelist = append(whitelist, whitelist_param)
	}

	var urls []string

	if list != "" && !custom_filter { // get file descriptor from given file or stdin
		f, err := os.Open(list)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			urls = append(urls, scanner.Text())
		}

	} else if stdin && !custom_filter {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			urls = append(urls, scanner.Text())
		}

	} else if (stdin || list != "") && custom_filter && url_based {
		urls = output_slice // if custom filter was provided and it's url based, set urls value to each each url of output
	}

	final_urls := make(map[string]string)
	var filtered_urls []string

	for _, uri := range urls { // iterate over urls
		uri = strings.TrimSpace(uri)
		if uri == "" {
			continue
		}

		u, _ := url.Parse(uri) // parse each url
		if u == nil {
			continue
		}

		// first of all check default filters if they were especified
		if strings.Contains(filter, "nocontent") { // enter here if "nocontent" filter is especified
			// check if any url dir is in text_content array
			text_found := 0
			for _, t := range strings.Split(u.Path, "/") { // iterate over any dir in url (i.e. /blog/gallery/images/)

				var builder strings.Builder
				for _, i := range t {
					if !unicode.IsDigit(i) { // remove all numbers from string (i.e. /review01/ --> /review/)
						builder.WriteRune(i)
					}
				}
				t2 := builder.String()

				// check if current iteration is in text_content array so it probably contains human content (useless)
				for _, text := range text_content {
					if t2 == text { // check match and if so, break loop
						text_found = 1
						break
					}
				}

				if text_found == 1 { // if match success, break again
					break
				}
			}

			if text_found == 1 { // finally, if match success continue to next iteration since the current urls isn't interesting
				continue
			}
		}

		if strings.Contains(filter, "hasparam") { // enter here if "hasparams" filter is especified
			if len(u.RawQuery) == 0 {
				continue
			}
		}

		if strings.Contains(filter, "noparam") { // enter here if "noparams" filter is especified
			if len(u.RawQuery) != 0 {
				continue
			}
		}

		pos := strings.LastIndexByte(u.Path, '.')
		// if url has no extension and "hasextension" filter is provided, skip to next iteration
		if (pos == -1) && (strings.Contains(filter, "hasextension")) {
			continue

			// if url has extension and "noextension" filter is provided, skip to next iteration
		} else if (pos != -1) && (strings.Contains(filter, "noextension")) {
			continue
		}

		extension := strings.ToLower(u.Path[pos+1:]) // get extension (php, html, js, png ...)

		// check if last url directory is a common useless js or css file (i.e. /js/chunk-7f801243.858.js)
		last_dir := strings.Split(u.Path, "/")[len(strings.Split(u.Path, "/"))-1]
		if extension == "js" && (strings.HasPrefix(last_dir, "chunk-") || strings.HasPrefix(last_dir, "app.") || strings.HasSuffix(last_dir, ".min.js")) {
			continue
		}

		// now create path regex-based patterns to exclude number-based urls
		// i.e. http://example.com/dir/11/index.html and http://example.com/dir/06/index.html
		pattern := createPattern(u.Path) // create pattern

		if !patternExists(pattern) { // check if current pattern already has been created and added to array
			seen_patterns = append(seen_patterns, pattern) // save pattern for later so integer based urls can be removed as expected

		} else if checkPattern(u.Path) && (len(u.Query()) == 0) { // check if pattern matches with path
			continue
		}

		if pos != -1 { // enter here if url has extension
			if len(blacklist) != 0 { // check if blacklists array has value
				if checkExtension(extension, blacklist) { // if function returns true, skip this url (blacklisted extension was found)
					continue
				}

				if checkExtension(extension, useless_extensions) { // also exclude urls with default extensions (images, audio, css...)
					continue
				}

			} else if len(whitelist) != 0 { // check if whitelist array has value
				if !checkExtension(extension, whitelist) { // if function returns false means that extension isn't in whitelist array (skip this url)
					continue
				}

			} else if checkExtension(extension, useless_extensions) { // if it's a "useless" extension continue over next url
				continue
			}
		}

		// at this point all the extensions checks are done (blacklist, whitelist and extensionless urls)
		// now start processing url params and applying "vuln" filter if is especified
		key := u.Host + u.Path

		if len(u.RawQuery) > 0 { // enter here if url has parameters
			// get query parameters
			queryParams := make([]string, len(u.Query()))

			i := 0
			for k := range u.Query() { // iterate over parameter names (i.e. page, id, query)
				if params_str == "" {
					queryParams[i] = k
				} else {
					queryParams[i] = k + "=FUZZ"
				}

				i++
			}
			//sort.Strings(queryParams)

			// at this point if any of the query parameters aren't present on seen_params, add it
			if checkParams(queryParams) == true { // if iteration query parameters have been already seen, jump to next one
				continue
			}

			key += "?" + strings.Join(queryParams, "&")
		}

		val, ok := final_urls[key]
		if ok {
			if u.Scheme == "https" && strings.HasPrefix(val, "http:") {
				if params_str == "" {
					fmt.Println(uri) // print url
				} else {
					fmt.Println(u.Scheme + "://" + key)
				}
				filtered_urls = append(filtered_urls, uri)
				csv_info = append(csv_info, []string{uri, filters})

				counter += 1
			}
		} else {
			if params_str != "" {
				uri = u.Scheme + "://" + key
			}

			fmt.Println(uri)
			filtered_urls = append(filtered_urls, uri)
			csv_info = append(csv_info, []string{uri, filters})

			counter += 1
		}

		if output != "" { // if output is especified, write urls to file
			_, err = txt_out.WriteString(uri + "\n")
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	if json_output != "" {
		json_urls := FilterInfo{
			Urls:    filtered_urls,
			Filters: strings.Split(strings.TrimSpace(filter), ","),
			Length:  len(filtered_urls),
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

	if csv_output != "" {
		csv_out, err := os.Create(csv_output)
		if err != nil {
			log.Fatal(err)
		}

		writer := csv.NewWriter(csv_out)
		defer writer.Flush()

		headers := []string{"urls", "filters"}
		writer.Write(headers)
		for _, row := range csv_info {
			writer.Write(row)
		}
	}

	// finally add some logging to aid users
	if !quiet {
		if counter >= 1 {
			fmt.Println()
			if !custom_filter {
				core.Green(strconv.Itoa(counter)+" unique urls found", use_color)
			} else {
				core.Green(strconv.Itoa(counter)+" urls processed", use_color)
			}

			if output != "" {
				core.Green("Urls written to "+output+" (TXT)", use_color)
			}

			if json_output != "" {
				core.Green("Urls written to "+json_output+" (JSON)", use_color)
			}

			if csv_output != "" {
				core.Green("Urls written to "+csv_output+" (CSV)", use_color)
			}
		} else if counter == 0 && url_based {
			core.Red("No urls found", use_color)
		}

		if use_color {
			fmt.Println("["+green("+")+"] Elapsed time:", green(core.TimerDiff(t1)))
		} else {
			fmt.Println("[+] Elapsed time:", core.TimerDiff(t1))
		}
	}
}

/*

TLDR; Auxiliary functions

*/

func getPatternDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	path := filepath.Join(usr.HomeDir, ".config/go-recon/patterns/")
	_, err = os.Stat(path)
	if !os.IsNotExist(err) {
		return path, nil
	}

	return filepath.Join(usr.HomeDir, ".config/go-recon/patterns/"), nil
}

func listPatterns(quiet, use_color bool) {
	if !quiet {
		core.Magenta("Available patterns:", use_color)
		fmt.Println("\tDEFAULT:")
		fmt.Println("\t  nocontent: exclude urls that are likely to contain human content (i.e. blogs, stories, articles...)")
		fmt.Println("\t  hasparams: filter only for urls that have parameters (i.e. http://example.com/?p=123)")
		fmt.Println("\t  noparams: exclude urls that have parameters (i.e. http://example.com/blog)")
		fmt.Println("\t  hasextension: filter only for urls that have extensions (i.e. http://example.com/index.php)")
		fmt.Println("\t  noextension: exclude urls that have extensions (i.e. http://example.com/blog)\n")
	} else {
		fmt.Println("nocontent\nhasparams\nnoparams\nhasextension\nnoextension")
	}

	patterns_dir, err := getPatternDir()
	if err != nil {
		log.Fatal(err)
	}

	files, err := filepath.Glob(patterns_dir + "/*.json")
	if err != nil {
		log.Fatal(err)
	}

	if !quiet {
		fmt.Println("\tCUSTOM:")
	}
	for _, f := range files {
		pattern_name := f[len(patterns_dir)+1 : len(f)-5]

		f, err := os.Open(filepath.Join(patterns_dir, pattern_name+".json"))
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		p := PatternFormat{}
		dec := json.NewDecoder(f)
		err = dec.Decode(&p)
		if err != nil {
			log.Fatal(err)
		}

		if !quiet {
			if p.Description != "" {
				fmt.Println("\t  " + pattern_name + ": " + p.Description)
			} else {
				fmt.Println("\t  " + pattern_name)
			}
		} else {
			fmt.Println(pattern_name)
		}
	}
}

func isValidUrl(toTest string) bool {
	_, err := nurl.ParseRequestURI(toTest)
	if err != nil {
		return false
	}

	u, err := nurl.Parse(toTest)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}

func createPattern(path string) string {
	// creates patterns for URLs with integers in them
	newParts := []string{}
	parts := strings.Split(path, "/")

	for _, part := range parts {
		if isDigit(part) {
			newParts = append(newParts, `\d+`)
		} else {
			newParts = append(newParts, regexp.QuoteMeta(part))
		}
	}

	return strings.Join(newParts, "/")
}

func checkPattern(path string) bool {
	// checks if the URL matches any of the int patterns
	for _, p := range seen_patterns {
		matched, _ := regexp.MatchString(p, path)
		if matched {
			return true
		}
	}

	return false
}

func patternExists(pattern string) bool {
	// check if it has already been added to seen_patterns slice
	for _, p := range seen_patterns {
		if p == pattern {
			return true
		}
	}

	return false
}

func checkParams(params []string) bool {
	if len(seen_params) == 0 {
		for _, p := range params {
			seen_params = append(seen_params, p)
		}

		return false
	}

	param_length := len(params)
	counter := 0
	var params_to_append []string

	for _, i := range seen_params {
		//fmt.Println(i)
		for _, p := range params {
			//fmt.Println("  " + p)
			if i == p {
				//fmt.Println("Matched!")
				params_to_append = append(params_to_append, p)
				counter += 1
				continue
			}
		}
	}

	results := diffSlices(params, params_to_append)
	for _, r := range results {
		//fmt.Println("Appending", r, "to", seen_params)
		seen_params = append(seen_params, r)
	}

	//fmt.Println("Matched params:", counter, "of", param_length)
	if counter >= param_length {
		return true
	} else {
		return false
	}
}

func diffSlices(slice1 []string, slice2 []string) []string {
	// Create a map to store the elements of slice2 for efficient lookup
	slice2Map := make(map[string]bool)
	for _, str := range slice2 {
		slice2Map[str] = true
	}

	// Create a result slice to store strings not in slice2
	result := []string{}

	// Iterate through slice1 and add elements not present in slice2 to the result slice
	for _, str := range slice1 {
		if !slice2Map[str] {
			result = append(result, str)
		}
	}

	return result
}

func checkExtension(extension string, array_to_check []string) bool {
	for _, e := range array_to_check {
		if extension == e { // check if given extension is present on extensions list
			return true
		}
	}

	return false
}

// Extra auxiliary func

func isDigit(s string) bool {
	// checks if a string contains only digits
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}

	return len(s) > 0
}
