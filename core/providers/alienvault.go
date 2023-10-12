package providers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type alienvaultResponse struct {
	Detail     string `json:"detail"`
	Error      string `json:"error"`
	PassiveDNS []struct {
		Hostname string `json:"hostname"`
	} `json:"passive_dns"`
}

func AlienVault(domain string, results chan string, client *http.Client) error {
	res, err := client.Get(fmt.Sprintf("https://otx.alienvault.com/api/v1/indicators/domain/%s/passive_dns", domain))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		err = fmt.Errorf("HackerTarget is down or seems to be slow")
		return err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var response alienvaultResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	for _, record := range response.PassiveDNS {
		if (record.Hostname != "") && (!strings.HasPrefix(record.Hostname, "*.")) {
			results <- record.Hostname
		}
	}

	return nil
}
