package core

import (
	"crypto/tls"
	"net"
	"net/http"
	"strings"
  "math/rand"
	"time"
)

// Generate a random integer between range
func RandomInt(min int, max int) int {
	rand.Seed(time.Now().UnixNano())
	rand_int := rand.Intn(max-min+1) + min
	return rand_int
}

// Return random string based on an integer (length)
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}

	// Ensure the first character is not an integer
	for b[0] >= '0' && b[0] <= '9' {
		b[0] = charset[seededRand.Intn(len(charset))]
	}

	return string(b)
}

// function which aids users if they want to use
// a default client instance instead of creating a new one
func DefaultHttpClient() *http.Client {
	return CreateHttpClient(10000)
}

// create an http client with given timeout (in milliseconds),
// skip tls verify and some other useful settings
// don't follow redirects
// Example: client := CreateHttpClient(5000)
func CreateHttpClient(timeout int) *http.Client {
	// define timeout
	t := time.Duration(timeout) * time.Millisecond

	var transport = &http.Transport{
		Proxy:             http.ProxyFromEnvironment,
		MaxIdleConns:      30,
		IdleConnTimeout:   time.Second,
		DisableKeepAlives: true,
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, // Disable ssl verify
		DialContext: (&net.Dialer{
			Timeout:   t,
			KeepAlive: time.Second,
		}).DialContext,
	}

	redirect := func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse // Don't follow redirect
	}

	client := &http.Client{ // Create requests client
		Transport:     transport,
		CheckRedirect: redirect,
		Timeout:       t,
	}

	return client
}

// this functions does the same as CreateHttpClient() but this one follows redirects
func CreateHttpClientFollowRedirects(timeout int) *http.Client {
	// define timeout
	t := time.Duration(timeout) * time.Millisecond

	var transport = &http.Transport{
		Proxy:             http.ProxyFromEnvironment,
		MaxIdleConns:      30,
		IdleConnTimeout:   time.Second,
		DisableKeepAlives: true,
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, // Disable ssl verify
		DialContext: (&net.Dialer{
			Timeout:   t,
			KeepAlive: time.Second,
		}).DialContext,
	}

	client := &http.Client{ // Create requests client
		Transport: transport,
		Timeout:   t,
	}

	return client
}

// return current time for later chaining
// with TimerDiff() to get elapsed time
func StartTimer() time.Time {
	return time.Now()
}

// this function receives a time and
// returns difference between current time and given time
func TimerDiff(t1 time.Time) time.Duration {
	t2 := time.Now()
	diff := t2.Sub(t1)

	return diff.Round(10 * time.Millisecond)
}

func Version() string {
	return "v0.2"
}

func stringInSlice(str string, slice []string) bool {
	for _, i := range slice {
		if strings.ToLower(str) == strings.ToLower(i) {
			return true
		}
	}

	return false
}
