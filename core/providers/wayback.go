package providers

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

func Wayback(domain string, results chan string, client *http.Client) error {
	var found_subs []string
	var x int = 0

	res, err := client.Get(fmt.Sprintf("http://web.archive.org/cdx/search/cdx?url=*.%s/*&output=txt&fl=original&collapse=urlkey", domain))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		line, _ = url.QueryUnescape(line)

		extractor, err := regexp.Compile(`(?i)[a-zA-Z0-9\*_.-]+\.` + domain)
		if err != nil {
			return err
		}

		sub := extractor.FindString(line) // get subdomain by each line using regex
		if sub == "" {
			continue
		}

		sub = strings.ToLower(sub)
		sub = strings.TrimPrefix(sub, "25")
		sub = strings.TrimPrefix(sub, "2f")

		if !strings.HasPrefix(sub, "*.") { // verify that subdomain doesn't have wildcard at the beginning of string
			// get unique subdomains since this provider reports a lot of urls with same subdomains so they're already filtered for performance
			for _, f := range found_subs {
				if sub == f {
					x = 1
					break
				}
			}

			if x != 1 {
				found_subs = append(found_subs, sub)
				results <- sub
			}

			x = 0
		}
	}

	return nil
}
