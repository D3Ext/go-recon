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

func GetEndpoints(urls []string, results chan string, workers int, timeout int) {
	urls_c := make(chan string)

	// start workers
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)

		go func() {
			FetchEndpoints(urls_c, results, "", "", timeout)

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
	close(results)
	//out.Wait()

	return
}

func FetchEndpoints(urls <-chan string, results chan string, user_agent string, proxy string, timeout int) {
	for u := range urls {
		// Check if URL is valid
		if (!strings.Contains(u, ".")) || (u == "") {
			continue
		}

		var err error

		req, err := http.NewRequest("GET", u, nil)
		if err != nil {
			continue
		}

		if user_agent != "" {
			req.Header.Add("User-Agent", user_agent)
		}

		var c *http.Client
		if proxy == "" {
			c = CreateHttpClient(timeout)
		} else {
			c, err = CreateHttpClientWithProxy(timeout, proxy)
			if err != nil {
				continue
			}
		}

		resp, err := c.Do(req)
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
			return
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
}
