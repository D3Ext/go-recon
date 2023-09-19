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
	"regexp"
	"strings"
	"unicode"

	"github.com/D3Ext/go-recon/core"
)

var red func(a ...interface{}) string = color.New(color.FgRed).SprintFunc()
var cyan func(a ...interface{}) string = color.New(color.FgCyan).SprintFunc()
var green func(a ...interface{}) string = color.New(color.FgGreen).SprintFunc()
var magenta func(a ...interface{}) string = color.New(color.FgMagenta).SprintFunc()
var yellow func(a ...interface{}) string = color.New(color.FgYellow).SprintFunc()

type FilterInfo struct {
	Urls    []string `json:"urls"`
	Filters []string `json:"filters"`
	Length  int      `json:"length"`
}

func helpPanel() {
	fmt.Println(`Usage of gr-filter:
    -l)       file containing a list of urls to remove duplicates and useless ones (one domain per line)
    -b)       blacklisted extensions to exclude from filtered urls, default extensions are also excluded (separated by comma) (i.e. php,js,mp3)
    -w)       whitelisted extensions to filter for (separated by comma) (i.e. json,txt,php)
    -f)       custom filter to use (vuln, redirects, nocontent, hasparams, noparams, hasextension, noextension)
    -o)       file to write filtered urls into
    -oj)      file to write filtered urls into (JSON format)
    -c)       print colors on output
    -q)       don't print banner nor logging, only output
    -h)       print help panel

Examples:
    gr-filter -l urls.txt -o clean.txt -c
    gr-filter -l urls.txt -b html,asp
    gr-filter -l urls.txt -w json -x
    gr-filter -l urls.txt -f hasparams,nocontent -o param_urls.txt -q
    cat urls.txt | gr-filter -f vuln
    `)
}

var vuln_params = [190]string{"file", "document", "folder", "root", "path", "pg", "style", "pdf", "template", "php_path", "doc", "page", "name", "cat", "dir", "action", "board", "date", "detail", "download", "prefix", "include", "inc", "locate", "show", "site", "type", "view", "content", "layout", "mod", "conf", "daemon", "upload", "log", "ip", "cli", "cmd", "exec", "command", "execute", "ping", "query", "jump", "code", "reg", "do", "func", "arg", "option", "load", "process", "step", "read", "function", "req", "feature", "exe", "module", "payload", "run", "print", "callback", "checkout", "checkout_url", "continue", "data", "dest", "destination", "domain", "feed", "file_name", "file_url", "folder_url", "forward", "from_url", "go", "goto", "host", "html", "image_url", "img_url", "load_file", "load_url", "login_url", "logout", "navigation", "next", "next_page", "Open", "out", "page_url", "port", "redir", "redirect", "redirect_to", "redirect_uri", "redirect_url", "reference", "return", "return_path", "return_to", "returnTo", "return_url", "rt", "rurl", "target", "to", "uri", "url", "val", "validate", "window", "q", "s", "search", "lang", "keyword", "keywords", "year", "email", "p", "jsonp", "api_key", "api", "password", "emailto", "token", "username", "csrf_token", "unsubscribe_token", "id", "item", "page_id", "month", "immagine", "list_type", "terms", "categoryid", "key", "l", "begindate", "enddate", "select", "report", "role", "update", "user", "sort", "where", "params", "row", "table", "from", "sel", "results", "sleep", "fetch", "order", "column", "field", "delete", "string", "number", "filter", "access", "admin", "dbg", "debug", "edit", "grant", "test", "alter", "clone", "create", "disable", "enable", "make", "modify", "rename", "reset", "shell", "toggle", "adm", "cfg", "open", "img", "filename", "preview", "activity"}

var text_content = [111]string{"blog", "historias", "personal", "diario", "vida", "historia", "historias", "imagenes", "galeria", "consejos", "viajes", "experiencias", "prensa", "revista", "noticias", "articulos", "informacion", "opiniones", "comentarios", "novedades", "entrevistas", "actualidad", "cronicas", "reportajes", "reseñas", "editorial", "publicaciones", "textos", "escritos", "relatos", "comunicados", "analisis", "columnas", "temas", "contenidos", "lecturas", "blogspot", "sitio", "seccion", "archivo", "blogueros", "autores", "periodismo", "notas", "articulos-de-blog", "entrevistas-destacadas", "stories", "press", "magazine", "news", "articles", "opinions", "images", "comments", "updates", "interviews", "galery", "advices", "story", "stories", "current-affairs", "chronicles", "reports", "reviews", "life", "journal", "travel", "experiencies", "editorial", "publications", "texts", "writings", "tales", "announcements", "analysis", "columns", "topics", "section", "bloggers", "journalism", "notes", "blog-articles", "featured-interviews", "histoires", "presse", "magazine", "actualites", "articles", "information", "opinions", "commentaires", "mises-a-jour", "entretiens", "actualites", "chroniques", "reportages", "critiques", "editorial", "textes", "ecrits", "annonces", "analyse", "colonnes", "contenus", "lectures", "blogueurs", "auteurs", "journalisme", "notes", "articles-de-blog", "interviews-a-la-une"}

var redirects_params = [23]string{"url", "from_url", "load_url", "file_url", "page_url", "file_name", "page", "folder", "folder_url", "login_url", "img_url", "return_url", "return_to", "next", "redirect", "redirect_to", "logout", "checkout", "checkout_url", "goto", "next_page", "file", "load_file"}

var useless_extensions = []string{"png", "jpeg", "gif", "jpg", "pjpeg", "svg", "jfif", "avif", "webp", "ico", "tiff", "ttf", "woff", "mp3", "avi", "mov", "mpeg", "wav", "msv", "wv", "cda", "vox", "ogg", "css", "swf"}

var blacklist []string

var whitelist []string

var seen_patterns []string

var seen_params []string

func main() {
	var list string
	var blacklist_param string
	var whitelist_param string
	var extensionless bool
	var filter string
	var output string
	var json_output string
	var stdin bool
	var quiet bool
	var use_color bool
	var help bool

	flag.StringVar(&list, "l", "", "file containing a list of urls to remove duplicates and useless ones (one url per line)")
	flag.StringVar(&blacklist_param, "b", "", "blacklisted extensions to exclude from filtered urls, default extensions are also excluded (separated by comma) (i.e. php,js,mp3)")
	flag.StringVar(&whitelist_param, "w", "", "whitelisted extensions to filter for (separated by comma) (i.e. json,txt,php)")
	flag.BoolVar(&extensionless, "x", false, "remove extensionless urls (default=disabled)")
	flag.StringVar(&filter, "f", "", "custom filter to use (vuln, hasparams, noparams)")
	flag.StringVar(&output, "o", "", "file to write filtered urls into")
	flag.StringVar(&json_output, "oj", "", "file to write filtered urls into (JSON format)")
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

	var counter int
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

	// before doing anything, check if invalid filter or filters are provided
	if filter != "" {
		if strings.Contains(filter, "hasparam") && strings.Contains(filter, "noparam") {
			core.Red("[-] Invalid filters provided! hasparam and noparam can't be used at the same time", use_color)
			os.Exit(0)

		} else if strings.Contains(filter, "hasextension") && strings.Contains(filter, "noextension") {
			core.Red("Invalid filters provided! hasextension and noextension can't be used at the same time", use_color)
			os.Exit(0)

		} else if strings.Contains(filter, "vuln") && strings.Contains(filter, "noparam") {
			core.Red("Invalid filters provided! vuln and noparam can't be used at the same time", use_color)
			os.Exit(0)
		}

		if !strings.Contains(filter, "vuln") &&
			!strings.Contains(filter, "hasparam") &&
			!strings.Contains(filter, "noparam") &&
			!strings.Contains(filter, "hasextension") &&
			!strings.Contains(filter, "noextension") &&
			!strings.Contains(filter, "nocontent") &&
			!strings.Contains(filter, "redirect") {

			core.Red("Invalid filter/s provided!", use_color)
			os.Exit(0)
		}
	}

	if !quiet {
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

	var f *os.File
	if list != "" { // get file descriptor from given file or stdin
		f, err = os.Open(list)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

	} else if stdin {
		f = os.Stdin
	}

	var urls []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
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

		pos := strings.LastIndexByte(u.Path, '.')
		// if url has no extension and "hasextension" filter is provided, skip to next iteration
		if (pos == -1) && (strings.Contains(filter, "hasextension")) {
			continue

			// if url has extension and "noextension" filter is provided, skip to next iteration
		} else if (pos != -1) && (strings.Contains(filter, "noextension")) {
			continue

			// if url has no parameters and "vuln" filter is provided, skip to next iteration
		} else if (len(u.RawQuery) == 0) && (strings.Contains(filter, "vuln")) {
			continue

			// if url has no parameters and "redirects" filter is provided, skip to next  iteration
		} else if (len(u.RawQuery) == 0) && (strings.Contains(filter, "redirect")) {
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

		extension := strings.ToLower(u.Path[pos+1:]) // get extension (.php, .html, .js, .png ...)

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
		// now check if filters were especified and if so, apply them

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

		if strings.Contains(filter, "redirect") { // enter here if "redirect" filter is especified
			if len(u.RawQuery) == 0 {
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

		if strings.Contains(filter, "vuln") {
			if len(u.RawQuery) == 0 {
				continue
			}
		}

		// now start processing url params and applying "vuln" filter if is especified
		key := u.Host + u.Path
		var vuln_found int

		if len(u.RawQuery) > 0 { // enter here if url has parameters
			// get query parameters
			queryParams := make([]string, len(u.Query()))

			i := 0
			for k := range u.Query() { // iterate over parameter names (i.e. page, id, query)
				// check vuln filter
				if strings.Contains(filter, "vuln") { // perform logic if filter is especified
					for _, v := range vuln_params {
						if v == k { // potential vuln param
							queryParams[i] = k
							vuln_found = 1
							break
						}
					}

					if vuln_found == 1 {
						break
					}

				} else if strings.Contains(filter, "redirect") {
					for _, v := range redirects_params {
						if v == k { // potential redirects param
							queryParams[i] = k
							vuln_found = 1
							break
						}
					}

					if vuln_found == 1 {
						break
					}

				} else {
					queryParams[i] = k
				}

				i++
			}
			//sort.Strings(queryParams)

			if (strings.Contains(filter, "vuln")) && (vuln_found == 1) { // check if vuln parameter was found
				for _, p := range queryParams {
					seen_params = append(seen_params, p)
				}

			} else if (strings.Contains(filter, "redirect")) && (vuln_found == 1) {
				for _, p := range queryParams {
					seen_params = append(seen_params, p)
				}

			} else if !checkParams(queryParams) { // if iteration query parameters have been already seen, jump to next one
				for _, p := range queryParams {
					seen_params = append(seen_params, p)
				}

			} else {
				continue
			}

			key += "?" + strings.Join(queryParams, "&")
		}

		/*// now create path regex-based patterns to exclude number-based urls
		  // i.e. http://example.com/dir/11/index.html and http://example.com/dir/06/index.html
		  pattern := createPattern(u.Path) // create pattern

		  if !patternExists(pattern) { // check if current pattern already has been created and added to array
		    seen_patterns = append(seen_patterns, pattern) // save pattern for later so integer based urls can be removed as expected
		    //fmt.Println(pattern, "-", seen_patterns)

		  } else if checkPattern(u.Path) && more_params != true { // check if pattern matches with path
		    //fmt.Println(pattern, "-", seen_patterns)
		    continue
		  }*/

		if (strings.Contains(filter, "vuln")) || (strings.Contains(filter, "redirect")) { // ensure that non vulnerable urls are excluded
			if vuln_found != 1 {
				continue
			}
		}

		val, ok := final_urls[key]
		if ok {
			if u.Scheme == "https" && strings.HasPrefix(val, "http:") {
				fmt.Println(uri) // print url
				filtered_urls = append(filtered_urls, uri)

				if output != "" { // if output is especified, write urls to file
					_, err = out_f.WriteString(uri + "\n")
					if err != nil {
						log.Fatal(err)
					}
				}

				counter += 1
			}
		} else {
			fmt.Println(uri) // print url
			filtered_urls = append(filtered_urls, uri)

			if output != "" { // if output is especified, write urls to file
				_, err = out_f.WriteString(uri + "\n")
				if err != nil {
					log.Fatal(err)
				}
			}
			counter += 1
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

		_, err = out_f.WriteString(string(json_body))
		if err != nil {
			log.Fatal(err)
		}
	}

	if counter == 0 { // check if no url was filtered
		core.Red("No urls found", use_color)
	}

	// finally add some logging to aid users
	if !quiet {
		if counter >= 1 {
			if use_color {
				fmt.Println("\n["+green("+")+"]", counter, "unique urls found")
			} else {
				fmt.Println("\n[+]", counter, "unique urls found")
			}

			if output != "" {
				core.Green("Urls written to "+output, use_color)
			} else if json_output != "" {
				core.Green("Urls written to "+json_output, use_color)
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

/*

TLDR; Auxiliary functions

*/

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
		return false
	}

	if len(params) > len(seen_params) { // if it has more params than already seen, it's a unique url
		return false
	}

	param_length := len(params)
	counter := 0

	for _, i := range seen_params {
		for _, p := range params {
			if i == p {
				counter += 1
				//return true
			}
		}
	}

	if counter == param_length {
		return true
	} else {
		return false
	}
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
