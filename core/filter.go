package core

import (
	"net/url"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

var vuln_params []string = []string{"file", "document", "folder", "root", "path", "pg", "style", "pdf", "template", "php_path", "doc", "page", "name", "cat", "dir", "action", "board", "date", "detail", "download", "prefix", "include", "inc", "locate", "show", "site", "type", "view", "content", "layout", "mod", "conf", "daemon", "upload", "log", "ip", "cli", "cmd", "exec", "command", "execute", "ping", "query", "jump", "code", "reg", "do", "func", "arg", "option", "load", "process", "step", "read", "function", "req", "feature", "exe", "module", "payload", "run", "print", "callback", "checkout", "checkout_url", "continue", "data", "dest", "destination", "domain", "feed", "file_name", "file_url", "folder_url", "forward", "from_url", "go", "goto", "host", "html", "image_url", "img_url", "load_file", "load_url", "login_url", "logout", "navigation", "next", "next_page", "Open", "out", "page_url", "port", "redir", "redirect", "redirect_to", "redirect_uri", "redirect_url", "reference", "return", "return_path", "return_to", "returnTo", "return_url", "rt", "rurl", "target", "to", "uri", "url", "val", "validate", "window", "q", "s", "search", "lang", "keyword", "keywords", "year", "email", "p", "jsonp", "api_key", "api", "password", "emailto", "token", "username", "csrf_token", "unsubscribe_token", "id", "item", "page_id", "month", "immagine", "list_type", "terms", "categoryid", "key", "l", "begindate", "enddate", "select", "report", "role", "update", "user", "sort", "where", "params", "row", "table", "from", "sel", "results", "sleep", "fetch", "order", "column", "field", "delete", "string", "number", "filter", "access", "admin", "dbg", "debug", "edit", "grant", "test", "alter", "clone", "create", "disable", "enable", "make", "modify", "rename", "reset", "shell", "toggle", "adm", "cfg", "open", "img", "filename", "preview", "activity"}

var text_content []string = []string{"blog", "historias", "personal", "diario", "vida", "historia", "historias", "imagenes", "galeria", "consejos", "viajes", "experiencias", "prensa", "revista", "noticias", "articulos", "informacion", "opiniones", "comentarios", "novedades", "entrevistas", "actualidad", "cronicas", "reportajes", "reseñas", "editorial", "publicaciones", "textos", "escritos", "relatos", "comunicados", "analisis", "columnas", "temas", "contenidos", "lecturas", "blogspot", "sitio", "seccion", "archivo", "blogueros", "autores", "periodismo", "notas", "articulos-de-blog", "entrevistas-destacadas", "stories", "press", "magazine", "news", "articles", "opinions", "images", "comments", "updates", "interviews", "galery", "advices", "story", "stories", "current-affairs", "chronicles", "reports", "reviews", "life", "journal", "travel", "experiencies", "editorial", "publications", "texts", "writings", "tales", "announcements", "analysis", "columns", "topics", "section", "bloggers", "journalism", "notes", "blog-articles", "featured-interviews", "histoires", "presse", "magazine", "actualites", "articles", "information", "opinions", "commentaires", "mises-a-jour", "entretiens", "actualites", "chroniques", "reportages", "critiques", "editorial", "textes", "ecrits", "annonces", "analyse", "colonnes", "contenus", "lectures", "blogueurs", "auteurs", "journalisme", "notes", "articles-de-blog", "interviews-a-la-une"}

var useless_extensions []string = []string{"png", "jpeg", "gif", "jpg", "pjpeg", "pjp", "svg", "jfif", "avif", "webp", "ico", "bmp", "cur", "tif", "tiff", "eot", "ttf", "woff", "woff2", "mp3", "mp4", "avi", "mov", "wmv", "flv", "mkv", "webm", "mpg", "mpeg", "wav", "ogv", "gifv", "mng", "aa", "aac", "act", "aiff", "gsm", "m4a", "movpkg", "mmf", "msv", "wv", "cda", "vox", "ogg", "css", "swf"}

var blacklist []string

var whitelist []string

var seen_patterns []string

var seen_params []string

func FilterUrls(urls []string, filters []string) []string {

	var urls_to_return []string

	var vuln bool
	var hasparams bool
	var noparams bool
	var hasextension bool
	var noextension bool
	var nocontent bool

	for _, f := range filters { // check provided filters
		if f == "vuln" {
			vuln = true
		} else if (f == "hasparams") || (f == "hasparam") {
			hasparams = true
		} else if (f == "noparams") || (f == "noparam") {
			noparams = true
		} else if f == "hasextension" {
			hasextension = true
		} else if f == "noextension" {
			noextension = true
		} else if f == "nocontent" {
			nocontent = true
		}
	}

	final_urls := make(map[string]string)
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
		if (pos == -1) && (hasextension) {
			continue

			// if url has extension and "noextension" filter is provided, skip to next iteration
		} else if (pos != -1) && (noextension) {
			continue
		}

		extension := strings.ToLower(u.Path[pos+1:]) // get extension (.php, .html, .js, .png ...)

		if pos != -1 { // enter here if url has extension
			if checkExtension(extension, useless_extensions) { // if it's a "useless" extension continue over next url
				continue
			}
		}

		// at this point all the extensions checks are done (blacklist, whitelist and extensionless urls)
		// now check if filters were especified and if so, apply them

		if nocontent { // enter here if "nocontent" filter is especified
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

				for _, text := range text_content { // check if current iteration is in text_content array so it probably contains personal texts (useless)
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

		if hasparams { // enter here if "hasparams" filter is especified
			if len(u.RawQuery) == 0 {
				continue
			}
		}

		if noparams { // enter here if "noparams" filter is especified
			if len(u.RawQuery) != 0 {
				continue
			}
		}

		// now start processing url params and applying "vuln" filter if is especified
		key := u.Host + u.Path
		var more_params bool

		if (len(u.RawQuery) == 0) && (vuln) { // continue to next iteration since "vuln" filter needs parameters
			continue

		} else if len(u.RawQuery) > 0 { // enter here if url has parameters
			// get query parameters
			queryParams := make([]string, len(u.Query()))

			i := 0
			vuln_found := 0
			for k := range u.Query() { // iterate over parameter names (i.e. page, id, query)
				// check vuln filter
				if vuln { // perform logic if filter is especified
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

				} else {
					queryParams[i] = k
				}

				i++
			}
			sort.Strings(queryParams)

			if checkParams(queryParams) {
				continue
			}

			if (vuln) && (vuln_found == 1) { // check if vuln parameter was found
				more_params = true // set variable to true so current urls isn't excluded later
				for _, p := range queryParams {
					seen_params = append(seen_params, p)
				}

			} else if !checkParams(queryParams) { // if iteration query parameters have been already seen, jump to next one
				more_params = true
				for _, p := range queryParams {
					seen_params = append(seen_params, p)
				}

			} else {
				continue
			}

			key += "?" + strings.Join(queryParams, "&")
		}

		// now create path regex-based patterns to exclude number-based urls
		// i.e. http://example.com/dir/11/index.html and http://example.com/dir/06/index.html
		pattern := createPattern(u.Path) // create pattern

		if !patternExists(pattern) { // check if current pattern already has been created and added to array
			seen_patterns = append(seen_patterns, pattern) // save pattern for later so integer based urls can be removed as expected

		} else if checkPattern(u.Path) && more_params != true { // check if pattern matches with path
			continue
		}

		val, ok := final_urls[key]
		if ok {
			if u.Scheme == "https" && strings.HasPrefix(val, "http:") {
				urls_to_return = append(urls_to_return, uri) // save url
			}
		} else {
			urls_to_return = append(urls_to_return, uri) // save url
		}
	}

	return urls_to_return
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
		return false
	}

	if len(seen_params) != len(params) { // if it has more params than already seen, it's a unique url
		return false
	}

	for _, i := range seen_params {
		for _, p := range params {
			if i == p {
				return true
			}
		}
	}

	return false
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
