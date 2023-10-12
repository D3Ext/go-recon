package providers

import (
	"bufio"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

func Digitorus(domain string, results chan string, client *http.Client) error {
	res, err := client.Get(fmt.Sprintf("https://certificatedetails.com/%s", domain))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		extractor, err := regexp.Compile(`(?i)[a-zA-Z0-9\*_.-]+\.` + domain)
		if err != nil {
			return err
		}
		subdomains := extractor.FindAllString(line, -1)

		for _, sub := range subdomains {
			if (sub != "") && (!strings.HasPrefix(sub, "*.")) {
				results <- sub
			}
		}
	}

	return nil
}
