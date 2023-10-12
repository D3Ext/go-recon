package providers

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

func WaybackUrls(domain string, results chan string, client *http.Client, workers int, recursive bool) error {
	// first we get the max number of pages
	var max_page int
	var err error
	var page_url string

	if !recursive {
		page_url = "http://web.archive.org/cdx/search/cdx?url=" + domain + "/*&showNumPages=true"
	} else {
		page_url = "http://web.archive.org/cdx/search/cdx?url=*." + domain + "/*&showNumPages=true"
	}

	req, err := http.NewRequest("GET", page_url, nil)
	if err != nil {
		return errors.New("WaybackMachine seems to be slow or down!")
	}

	res, err := client.Do(req)
	if err != nil {
		return errors.New("WaybackMachine seems to be slow or down!")
	}
	defer res.Body.Close()

	raw, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return errors.New("WaybackMachine seems to be slow or down!")
	}

	max_page, err = strconv.Atoi(strings.TrimSuffix(string(raw), "\n"))
	if err != nil {
		return errors.New("invalid Wayback page number")
	}

	var rangeSize int // split number of pages between workers to get the total number of iterations

	if max_page%workers == 0 && max_page != 2 { // check if workers is multiple of max_page
		rangeSize = max_page / workers

	} else if (max_page%workers != 0) && (workers > 1) { // if not, subtract 1 or add 5 as maximum to get a
		for i := -1; i <= 5; i++ {
			if max_page%(workers+i) == 0 {
				workers = workers + i
				break
			}
		}

		rangeSize = max_page / workers

	} else if (max_page >= 1) && (max_page <= 6) {
		rangeSize = 10
		workers = 1
	}

	var w sync.WaitGroup

	for i := 0; i < workers; i++ { // create n workers
		w.Add(1)

		go worker(domain, results, client, recursive, rangeSize*i, rangeSize*(i+1), &w)
	}

	w.Wait()

	return nil
}

func worker(domain string, results chan string, client *http.Client, recursive bool, start int, end int, wg *sync.WaitGroup) error {
	var current_url string

	for i := start; i < end; i++ {

		if !recursive {
			current_url = "http://web.archive.org/cdx/search/cdx?url=" + domain + "/*&output=json&collapse=urlkey&page=" + strconv.Itoa(i)
		} else {
			current_url = "http://web.archive.org/cdx/search/cdx?url=*." + domain + "/*&output=json&collapse=urlkey&page=" + strconv.Itoa(i)
		}

		req, err := http.NewRequest("GET", current_url, nil)
		if err != nil {
			continue
		}

		res, err := client.Do(req)
		if err != nil {
			continue
		}
		defer res.Body.Close()

		raw, err := ioutil.ReadAll(res.Body)
		if err != nil {
			continue
		}

		var wrapper [][]string
		err = json.Unmarshal(raw, &wrapper)
		if err != nil {
			continue
		}

		skip := true
		for _, urls := range wrapper {
			if skip {
				skip = false
				continue
			}

			results <- urls[2]
		}
	}

	wg.Done()

	return nil
}
