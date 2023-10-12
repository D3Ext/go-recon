package core

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

// main function to extract JS endpoints from a list of urls
// it receives a custom client for further customization
// Example: go gorecon.GetEndpointsFromFile(urls, results, 15, gorecon.DefaultClient())
func GetEndpoints(urls []string, results chan string, workers int, client *http.Client) error {
	urls_c := make(chan string)

	// start workers
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)

		go func() {
			FetchEndpoints(urls_c, results, client)

			wg.Done()
		}()
	}

	for _, u := range urls {
		if u != "" {
			urls_c <- u
		}
	}

	close(urls_c)
	wg.Wait()

	return nil
}

// this function receives urls from channel so
// it's better for concurrency and configuration
func FetchEndpoints(urls <-chan string, results chan string, client *http.Client) error { // nolint: gocyclo
	for u := range urls {
		// check if URL is valid
		if (!strings.Contains(u, ".")) || (u == "") {
			continue
		}

		var err error

		req, err := http.NewRequest("GET", u, nil)
		if err != nil {
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			continue
		}

		u, err := url.Parse(u)
		if err != nil {
			return err
		}

		doc.Find("script").Each(func(index int, s *goquery.Selection) {
			js, _ := s.Attr("src")

			if js != "" {
				if strings.HasPrefix(js, "http://") || strings.HasPrefix(js, "https://") {
					results <- js
				} else if strings.HasPrefix(js, "//") {
					js := fmt.Sprintf("%s:%s", u.Scheme, js)
					results <- js
				} else if strings.HasPrefix(js, "/") {
					js := fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, js)
					results <- js
				} else {
					js := fmt.Sprintf("%s://%s/%s", u.Scheme, u.Host, js)
					results <- js
				}
			}

			r := regexp.MustCompile(`[(\w./:)]*js`)
			matches := r.FindAllString(s.Contents().Text(), -1)
			for _, js := range matches {
				if strings.HasPrefix(js, "//") {
					js := fmt.Sprintf("%s:%s", u.Scheme, js)
					results <- js
				} else if strings.HasPrefix(js, "/") {
					js := fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, js)
					results <- js
				}
			}
		})

		doc.Find("div").Each(func(index int, s *goquery.Selection) {
			js, _ := s.Attr("data-script-src")
			if js != "" {
				if strings.HasPrefix(js, "http://") || strings.HasPrefix(js, "https://") {
					results <- js
				} else if strings.HasPrefix(js, "//") {
					js := fmt.Sprintf("%s:%s", u.Scheme, js)
					results <- js
				} else if strings.HasPrefix(js, "/") {
					js := fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, js)
					results <- js
				} else {
					js := fmt.Sprintf("%s://%s/%s", u.Scheme, u.Host, js)
					results <- js
				}
			}
		})
	}

	return nil
}
