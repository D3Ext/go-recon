package providers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type otxResult struct {
	HasNext    bool `json:"has_next"`
	ActualSize int  `json:"actual_size"`
	URLList    []struct {
		Domain   string `json:"domain"`
		URL      string `json:"url"`
		Hostname string `json:"hostname"`
		HTTPCode int    `json:"httpcode"`
		PageNum  int    `json:"page_num"`
		FullSize int    `json:"full_size"`
		Paged    bool   `json:"paged"`
	} `json:"url_list"`
}

func AlienVaultUrls(domain string, results chan string, client *http.Client, recursive bool) error {
	base_url := "https://otx.alienvault.com/api/v1/indicators/domain/" + domain + "/url_list?limit=100&page="

	for i := 1; i < 100; i++ {
		req, err := http.NewRequest("GET", base_url+strconv.Itoa(i), nil)
		if err != nil {
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		raw, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var result otxResult
		err = json.Unmarshal(raw, &result)
		if err != nil {
			return err
		}

		if result.HasNext == false {
			break
		}

		for _, e := range result.URLList {

			if recursive == false {
				current_url, err := url.Parse(e.URL)
				if err != nil {
					continue
				}
				hostname := strings.TrimPrefix(current_url.Hostname(), "www.")

				if hostname == domain {
					results <- e.URL
				}

			} else {
				results <- e.URL
			}
		}
	}

	return nil
}
