# Examples

Here is a list of the accessible function through the official API and some example files to see how the functions work 

## Subdomains

```go
// this function sents through provided channel all the gathered subdomains
// providers slice is used to configure the providers to use
// it also receives a client so you can custom most of the process
// Example: err := GetSubdomains("example.com", results, []string{"alienvault", "crt", "rapiddns", "wayback"}, gorecon.DefaultClient())
func GetSubdomains(domain string, results chan string, providers []string, client *http.Client) error {
	return core.GetSubdomains(domain, results, providers, client)
}

/*

providers functions

*/

func AlienVault(domain string, results chan string, client *http.Client) error {
  return p.AlienVault(domain, results, client)
}

func Anubis(domain string, results chan string, client *http.Client) error {
  return p.Anubis(domain, results, client)
}

func CommonCrawl(domain string, results chan string, client *http.Client) error {
  return p.CommonCrawl(domain, results, client)
}

func Crt(domain string, results chan string, client *http.Client) error {
  return p.Crt(domain, results, client)
}

func Digitorus(domain string, results chan string, client *http.Client) error {
  return p.Digitorus(domain, results, client)
}

func HackerTarget(domain string, results chan string, client *http.Client) error {
  return p.HackerTarget(domain, results, client)
}

func RapidDns(domain string, results chan string, client *http.Client) error {
  return p.RapidDns(domain, results, client)
}

func Wayback(domain string, results chan string, client *http.Client) error {
  return p.Wayback(domain, results, client)
}
```

## Secrets

```go
// this function receives a url and a client to look for
// potential leaked secrets like API keys (using regex)
// Example: secrets, err := gorecon.FindSecrets("http://github.com", gorecon.DefaultClient())
func FindSecrets(url string, client *http.Client) ([]string, error) {
	return core.FindSecrets(url, client)
}
```

## Filter

```go
// remove useless urls, duplicates and more
// to optimize results as much as possible from 
// a list of urls
// Example: new_urls := gorecon.FilterUrls(urls, []string{"hasparams"})
func FilterUrls(urls []string, filters []string) []string {
  return core.FilterUrls(urls, filters)
}
```

## Tech

```go
// this function send a request to given url and returns running technologies
// Example: techs, err := GetTech("http://github.com", gorecon.DefaultClient())
func GetTech(url string, timeout int) (map[string]struct{}, error) {
  return core.GetTech(url, timeout)
}
```

## 403

```go
// try different ways to bypass 403 status code urls
// returns slice of urls with payloads on them,
// a slice with their respective status codes, and
// finally an error
func Check403(url, word string, timeout int) ([]string, []int, error) {
  return core.Check403(url, word, timeout)
}
```

## WAF

```go
// this function send a request to url with an LFI payload
// to try to trigger the possible WAF (Web Application Firewall) i.e. Cloudflare
// Example: waf, err := gorecon.DetectWaf(url, "", "", gorecon.DefaultHttpClient())
func DetectWaf(url string, payload string, keyword string, timeout int) (string, error) {
  return core.DetectWaf(url, payload, keyword, timeout)
}
```

## AWS

```go
// this function returns all defined permutations for
// S3 buckets name generation
func GetAllPerms() []string {
	return core.GetAllPerms()
}

// this function returns more or less permutations based on given level
// 1 returns less permutations than 6 (1 lower, 5 higher)
func GetPerms(level int) []string {
	return core.GetPerms(level)
}
```

## DNS

```go
/*

type DnsInfo struct {
  Domain      string    // given domain
  CNAME       string    // returns the canonical name for the given host
  TXT         []string  // returns the DNS TXT records for the given domain name
  MX          []*net.MX //
  NS          []*net.NS //
  Hosts       []string  // returns a slice of given host's IPv4 and IPv6 addresses
}

*/

// main function for DNS information gathering
// it receives a domain and tries to find most important info
// and returns a DnsInfo struct and an error
func Dns(domain string) (core.DnsInfo, error) {
  return core.Dns(domain)
}
```

## Whois

```go
// send WHOIS query to given domain to retrieve public info
// Example: info, err := gorecon.Whois("hackthebox.com")
func Whois(domain string) (wp.WhoisInfo, error) {
  return core.Whois(domain)
}

/*

type WhoisInfo struct {
	Domain         *Domain  `json:"domain,omitempty"`
	Registrar      *Contact `json:"registrar,omitempty"`
	Registrant     *Contact `json:"registrant,omitempty"`
	Administrative *Contact `json:"administrative,omitempty"`
	Technical      *Contact `json:"technical,omitempty"`
	Billing        *Contact `json:"billing,omitempty"`
}

type Domain struct {
	ID                   string     `json:"id,omitempty"`
	Domain               string     `json:"domain,omitempty"`
	Punycode             string     `json:"punycode,omitempty"`
	Name                 string     `json:"name,omitempty"`
	Extension            string     `json:"extension,omitempty"`
	WhoisServer          string     `json:"whois_server,omitempty"`
	Status               []string   `json:"status,omitempty"`
	NameServers          []string   `json:"name_servers,omitempty"`
	DNSSec               bool       `json:"dnssec,omitempty"`
	CreatedDate          string     `json:"created_date,omitempty"`
	CreatedDateInTime    *time.Time `json:"created_date_in_time,omitempty"`
	UpdatedDate          string     `json:"updated_date,omitempty"`
	UpdatedDateInTime    *time.Time `json:"updated_date_in_time,omitempty"`
	ExpirationDate       string     `json:"expiration_date,omitempty"`
	ExpirationDateInTime *time.Time `json:"expiration_date_in_time,omitempty"`
}

type Contact struct {
	ID           string `json:"id,omitempty"`
	Name         string `json:"name,omitempty"`
	Organization string `json:"organization,omitempty"`
	Street       string `json:"street,omitempty"`
	City         string `json:"city,omitempty"`
	Province     string `json:"province,omitempty"`
	PostalCode   string `json:"postal_code,omitempty"`
	Country      string `json:"country,omitempty"`
	Phone        string `json:"phone,omitempty"`
	PhoneExt     string `json:"phone_ext,omitempty"`
	Fax          string `json:"fax,omitempty"`
	FaxExt       string `json:"fax_ext,omitempty"`
	Email        string `json:"email,omitempty"`
	ReferralURL  string `json:"referral_url,omitempty"`
}

*/
```

## Urls

```go
// main function to enumerate urls about provided domain, urls are sent through channel
// set "recursive" to false if you don't want to get urls related to subdomains
func GetAllUrls(domain string, results chan string, client *http.Client, recursive bool) error {
  return core.GetAllUrls(domain, results, client, recursive)
}

/*

providers functions

*/

func WaybackUrls(domain string, results chan string, client *http.Client, workers int, recursive bool) error {
	return providers.WaybackUrls(domain, results, client, workers, recursive)
}

func AlienVaultUrls(domain string, results chan string, client *http.Client, recursive bool) (error) {
	return providers.AlienVaultUrls(domain, results, client, recursive)
}

func UrlScanUrls(domain string, results chan string, client *http.Client, recursive bool, apikey string) (error) {
	return providers.UrlScanUrls(domain, results, client, recursive, apikey)
}
```

## Open Redirects

```go
// this function checks if given url is vulnerable to open redirect with provided payloads
// if keyword has value, it will be replaced with payloads
// Example: vuln_urls, err := CheckRedirect("http://example.com/index.php?p=FUZZ", client, []string{"bing.com", "//bing.com"}, "FUZZ")
func CheckRedirect(url string, c *http.Client, payloads []string, keyword string) ([]string, error) {
	return core.CheckRedirect(url, c, payloads, keyword)
}

// return all defined payloads
func GetPayloads() []string {
	return core.GetPayloads()
}

// return common payloads
func GetCommonPayloads() []string {
  return core.GetCommonPayloads()
}
```

## JS

```go
// main function to extract JS endpoints from a list of urls
// it receives a custom client for further customization
// Example: go gorecon.GetEndpointsFromFile(urls, results, 15, gorecon.DefaultClient())
func GetEndpoints(urls []string, results chan string, workers int, client *http.Client) error {
	return core.GetEndpoints(urls, results, workers, client)
}

// this function receives urls from channel so 
// it's better for concurrency and configuration
func FetchEndpoints(urls <-chan string, results chan string, client *http.Client) error {
	return core.FetchEndpoints(urls, results, client)
}
```




