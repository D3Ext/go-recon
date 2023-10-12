package providers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type indexResponse struct {
	ID     string `json:"id"`
	APIURL string `json:"cdx-api"`
}

const (
	indexURL     = "https://index.commoncrawl.org/collinfo.json"
	maxYearsBack = 5
)

var year = time.Now().Year()

func CommonCrawl(domain string, results chan string, client *http.Client) error {
	res, err := client.Get(indexURL)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var indexes []indexResponse
	err = json.Unmarshal(body, &indexes)
	if err != nil {
		return err
	}

	years := make([]string, 0)
	for i := 0; i < maxYearsBack; i++ {
		years = append(years, strconv.Itoa(year-i))
	}

	searchIndexes := make(map[string]string)
	for _, year := range years {
		for _, index := range indexes {
			if strings.Contains(index.ID, year) {
				if _, ok := searchIndexes[year]; !ok {
					searchIndexes[year] = index.APIURL
					break
				}
			}
		}
	}

	for _, apiURL := range searchIndexes {
		res, err = client.Get(fmt.Sprintf("%s?url=*.%s", apiURL, domain))
		if err != nil {
			return err
		}

		scanner := bufio.NewScanner(res.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}

			line, _ = url.QueryUnescape(line)
			var subdomain string

			if line != "" {
				// fix for triple encoded URL
				subdomain = strings.ToLower(subdomain)
				subdomain = strings.TrimPrefix(subdomain, "25")
				subdomain = strings.TrimPrefix(subdomain, "2f")

				if (subdomain != "") && (!strings.HasPrefix(subdomain, "*.")) {
					results <- subdomain
				}
			}
		}
	}

	return nil
}
