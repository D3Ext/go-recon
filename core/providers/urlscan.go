package providers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type searchResults struct {
	Results []struct {
		Task struct {
			Visibility string    `json:"visibility"`
			Method     string    `json:"method"`
			Time       time.Time `json:"time"`
			Source     string    `json:"source"`
			URL        string    `json:"url"`
		} `json:"task"`
		Stats struct {
			UniqIPs           int `json:"uniqIPs"`
			ConsoleMsgs       int `json:"consoleMsgs"`
			DataLength        int `json:"dataLength"`
			EncodedDataLength int `json:"encodedDataLength"`
			Requests          int `json:"requests"`
		} `json:"stats"`
		Page struct {
			Country string `json:"country"`
			Server  string `json:"server"`
			City    string `json:"city"`
			Domain  string `json:"domain"`
			IP      string `json:"ip"`
			Asnname string `json:"asnname"`
			Asn     string `json:"asn"`
			URL     string `json:"url"`
			Ptr     string `json:"ptr"`
		} `json:"page"`
		UniqCountries int    `json:"uniq_countries"`
		ID            string `json:"_id"`
		Result        string `json:"result"`
	} `json:"results"`
	Total int `json:"total"`
}

func UrlScanUrls(domain string, results chan string, client *http.Client, recursive bool, apikey string) error {
	base_url := "https://urlscan.io/api/v1/search/?q=domain:" + domain + "&size=10000"

	req, err := http.NewRequest("GET", base_url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("API-Key", apikey)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var res searchResults

	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(raw, &res)
	if err != nil {
		return err
	}

	for _, e := range res.Results {

		if !recursive {
			current_url, err := url.Parse(e.Task.URL)
			if err != nil {
				continue
			}
			hostname := strings.TrimPrefix(current_url.Hostname(), "www.")

			if hostname == domain {
				results <- e.Task.URL
			}

		} else {
			results <- e.Task.URL
		}
	}

	return nil
}
