package core

import (
	wappalyzer "github.com/projectdiscovery/wappalyzergo"
	"io/ioutil"
	"net/http"
	"time"
)

func GetTech(url string, timeout int) (map[string]struct{}, error) {
	client := &http.Client{ // Create requests client
		Timeout: time.Duration(timeout) * time.Millisecond,
	}

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Connection", "close")
	req.Close = true

	resp, err := client.Do(req) // Send request
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	wappalyzer_client, err := wappalyzer.New()
	if err != nil {
		return nil, err
	}

	techs := wappalyzer_client.Fingerprint(resp.Header, data)
	return techs, nil
}
