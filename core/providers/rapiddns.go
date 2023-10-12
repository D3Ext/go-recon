package providers

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

func RapidDns(domain string, results chan string, client *http.Client) error {
	res, err := client.Get(fmt.Sprintf("https://rapiddns.io/subdomain/%s?full=1", domain))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	extractor, err := regexp.Compile(`(?i)[a-zA-Z0-9\*_.-]+\.` + domain)
	if err != nil {
		return err
	}
	subdomains := extractor.FindAllString(string(body), -1)

	for _, sub := range subdomains {
		if (sub != "") && (!strings.HasPrefix(sub, "*.")) {
			results <- sub
		}
	}

	return nil
}
