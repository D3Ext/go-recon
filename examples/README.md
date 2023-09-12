# Examples

Here is a list of the accessible function through the official API and some example files to see how the functions work 

## Subdomains

```go
// results := make(chan string)
// go gorecon.GetAllSubdomains("example.com", results, 10000)
func GetAllSubdomains(domain string, subdomains chan string, timeout int) {
  core.GetAllSubdomains(domain, subdomains, timeout)
}

// subdomains, err := gorecon.Crt("example.com", 8000)
func Crt(domain string, timeout int) ([]string, error) {
  return core.Crt(domain, timeout)
}

// subdomains, err := gorecon.HackerTarget("example.com", 8000)
func HackerTarget(domain string, timeout int) ([]string, error) {
  return core.HackerTarget(domain, timeout)
}

// subdomains, err := gorecon.AlienVault("example.com", 8000)
func AlienVault(domain string, timeout int) ([]string, error) {
  return core.AlienVault(domain, timeout)
}
```

## Secrets

```go
func FindSecrets(url string, timeout int) ([]string, error) {
  return core.FindSecrets(url, timeout)
}
```

## Filter

```go
func FilterUrls(urls []string, filters []string) []string {
  return core.FilterUrls(urls, filters)
}
```

## Tech

```go
func GetTech(url string, timeout int) (map[string]struct{}, error) {
  return core.GetTech(url, timeout)
}
```

## 403

```go
func Check403(url, word string, timeout int) ([]string, []int, error) {
  return core.Check403(url, word, timeout)
}
```

## WAF

```go
func DetectWaf(url string, payload string, keyword string, timeout int) (string, error) {
  return core.DetectWaf(url, payload, keyword, timeout)
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

func Dns(domain string) (core.DnsInfo, error) {
  return core.Dns(domain)
}
```

## Whois

```go
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
// results := make(chan string)
// go gorecon.GetAllUrls("example.com", results, 5000, true)
func GetAllUrls(domain string, results chan string, timeout int, recursive bool) {
  core.GetAllUrls(domain, results, timeout, recursive)
}

// results := make(chan string)
// go gorecon.GetAllUrls("example.com", results, 5000, true)
func GetWaybackUrls(domain string, results chan string, timeout int, recursive bool) error {
  return core.GetWaybackUrls(domain, results, timeout, recursive)
}

func GetOTXUrls(domain string, timeout int, recursive bool) ([]string, error) {
  return core.GetOTXUrls(domain, timeout, recursive)
}

func GetUrlScanUrls(domain string, timeout int, apikey string) ([]string, error) {
  return core.GetUrlScanUrls(domain, timeout, apikey)
}
```

## Open Redirects

```go
func CheckRedirect(url string, c *http.Client, payloads []string, keyword string) ([]string, error) {
  return core.CheckRedirect(url, c, payloads, keyword)
}

func GetPayloads() []string {
  return core.GetPayloads()
}
```

## JS

```go
func GetEndpoints(urls []string, results chan string, workers int, timeout int) {
  core.GetEndpoints(urls, results, workers, timeout)
}

func FetchEndpoints(urls <-chan string, results chan string, user_agent string, timeout int) {
  core.FetchEndpoints(urls, results, user_agent, timeout)
}
```



