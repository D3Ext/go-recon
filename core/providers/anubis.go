package providers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func Anubis(domain string, results chan string, client *http.Client) error {
	res, err := client.Get(fmt.Sprintf("https://jonlu.ca/anubis/subdomains/%s", domain))
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var subdomains []string
	err = json.Unmarshal(body, &subdomains)
	if err != nil {
		return err
	}

	for _, s := range subdomains {
		if (s != "") && (!strings.HasPrefix(s, "*.")) {
			results <- s
		}
	}

	return nil
}
