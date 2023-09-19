package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// Structs used by APIs
type CTLogs []*CTLog

type CTLog struct { // Custom struct for crt.sh api
	IssuerCaID        int    `json:"issuer_ca_id"`
	IssuerName        string `json:"issuer_name"`
	NameValue         string `json:"name_value"`
	MinCertID         int    `json:"min_cert_id"`
	MinEntryTimestamp string `json:"min_entry_timestamp"`
	NotBefore         string `json:"not_before"`
	NotAfter          string `json:"not_after"`
}

type AlienvaultResponse struct {
	PassiveDNS []struct {
		Hostname string `json:"hostname"`
	} `json:"passive_dns"`
}

// Define API URLs
const crtsh_url string = "https://crt.sh/?output=json&CN="
const hackertarget_url string = "https://api.hackertarget.com/hostsearch/?q="
const alienvault_url string = "https://otx.alienvault.com/api/v1/indicators/domain/"

// GetAllSubdomains("example.com", subs_chan, 10000)
func GetAllSubdomains(dom string, c chan string, timeout int) error {
	var chan_subs []string

	// Get all subdomains
	subdomains1, err := HackerTarget(dom, timeout)
	if err == nil {
		//close(c)
		//return err

		for _, entry := range subdomains1 { // Add all subdomains from Crt.sh
			if (notInSlice(entry, chan_subs)) && (!strings.HasPrefix(entry, "*.")) { // Check if value already in chan_subs array
				c <- entry
			}

			chan_subs = append(chan_subs, entry)
		}
	}

	subdomains2, err := AlienVault(dom, timeout)
	if err == nil {
		//close(c)
		//return err

		for _, entry := range subdomains2 { // Add all subdomains from HackerTarget
			if (notInSlice(entry, chan_subs)) && (!strings.HasPrefix(entry, "*.")) { // Check if value already in chan_subs array
				c <- entry
			}

			chan_subs = append(chan_subs, entry)
		}
	}

	subdomains3, err := Crt(dom, timeout)
	if err == nil { //&& (strings.Contains(err.Error(), "context deadline exceeded") == false) {
		//close(c)
		//return err

		for _, entry := range subdomains3 { // Add subdomains from AlienVault
			if (notInSlice(entry, chan_subs)) && (!strings.HasPrefix(entry, "*.")) { // Check if value already in chan_subs array
				c <- entry
			}

			chan_subs = append(chan_subs, entry)
		}
	}

	// Close subdomains channel
	close(c)
	return nil
}

// Crt("example.com", 8000)
func Crt(dom string, timeout int) ([]string, error) {
	var crt_subdomains []string

	// Use crt.sh API
	req := http.Client{
		Timeout: time.Duration(timeout) * time.Millisecond,
	}

	res, err := req.Get(crtsh_url + dom)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		err = fmt.Errorf("crt.sh is down or seems to be slow")
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var logs CTLogs
	err = json.Unmarshal(body, &logs)
	if err != nil {
		return nil, err
	}

	for _, c := range logs {
		if (c.NameValue != "") && (c.NameValue != "\n") && (!strings.HasPrefix(c.NameValue, "*.")) && (c.NameValue != dom) {
			crt_subdomains = append(crt_subdomains, c.NameValue)
		}
	}

	// remove duplicates
	keys := make(map[string]bool)
	results := []string{}

	for _, entry := range crt_subdomains {
		if (entry == "") || (strings.Contains(entry, " ")) {
			continue
		}

		if _, value := keys[entry]; !value {
			keys[entry] = true
			results = append(results, entry)
		}
	}

	return results, nil
}

// HackerTarget("example.com", 10000)
func HackerTarget(dom string, timeout int) ([]string, error) {
	var hackertarget_subdomains []string

	// Use HackerTarget API
	req := http.Client{
		Timeout: time.Duration(timeout) * time.Millisecond,
	}

	res, err := req.Get(hackertarget_url + dom)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		err = fmt.Errorf("HackerTarget is down or seems to be slow")
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	for _, entry := range strings.Split(string(body), "\n") {
		sub := strings.Split(entry, ",")[0]
		if (sub != "") && (sub != "\n") && (!strings.Contains(sub, "API count exceeded")) && (sub != dom) && (!strings.HasPrefix(sub, "*.")) {
			hackertarget_subdomains = append(hackertarget_subdomains, sub)
		}
	}

	return hackertarget_subdomains, nil
}

// AlienVault("example.com", 8000)
func AlienVault(dom string, timeout int) ([]string, error) {
	var alienvault_subdomains []string

	// Use AlienVault API
	req := http.Client{
		Timeout: time.Duration(timeout) * time.Millisecond,
	}

	res, err := req.Get(alienvault_url + dom + "/passive_dns")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		err = fmt.Errorf("AlienVault is down or seems to be slow")
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var alien AlienvaultResponse
	err = json.Unmarshal(body, &alien)
	if err != nil {
		return nil, err
	}

	for _, entry := range alien.PassiveDNS {
		if (entry.Hostname != "") && (entry.Hostname != "\n") && (!strings.HasPrefix(entry.Hostname, "*.")) && (entry.Hostname != dom) {
			alienvault_subdomains = append(alienvault_subdomains, entry.Hostname)
		}
	}

	// remove duplicates
	keys := make(map[string]bool)
	results := []string{}

	for _, entry := range alienvault_subdomains {
		if (entry == "") || (strings.Contains(entry, " ")) {
			continue
		}

		if _, value := keys[entry]; !value {
			keys[entry] = true
			results = append(results, entry)
		}
	}

	return results, nil
}

func notInSlice(val string, slice_to_check []string) bool {
	for _, x := range slice_to_check {
		if x == val {
			return false
		}
	}

	return true
}
