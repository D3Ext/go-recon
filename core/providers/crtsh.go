package providers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type CTLogs []*CTLog

type CTLog struct {
	IssuerCaID        int    `json:"issuer_ca_id"`
	IssuerName        string `json:"issuer_name"`
	NameValue         string `json:"name_value"`
	MinCertID         int    `json:"min_cert_id"`
	MinEntryTimestamp string `json:"min_entry_timestamp"`
	NotBefore         string `json:"not_before"`
	NotAfter          string `json:"not_after"`
}

func Crt(domain string, results chan string, client *http.Client) error {
	res, err := client.Get(fmt.Sprintf("https://crt.sh/?q=%%25.%s&output=json", domain))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		err = fmt.Errorf("crt.sh is down or seems to be slow")
		return err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var logs CTLogs
	err = json.Unmarshal(body, &logs)
	if err != nil {
		return err
	}

	for _, c := range logs {
		for _, sub := range strings.Split(c.NameValue, "\n") {
			if (sub != "") && (!strings.HasPrefix(sub, "*.")) {
				results <- sub
			}
		}
	}

	return nil
}
