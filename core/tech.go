package core

import (
	wappalyzer "github.com/projectdiscovery/wappalyzergo"
	"io/ioutil"
	"net/http"
)

// this function send a request to given url and returns running technologies
// Example: techs, err := GetTech("http://github.com", gorecon.DefaultClient())
func GetTech(url string, client *http.Client) (map[string]struct{}, error) {

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
