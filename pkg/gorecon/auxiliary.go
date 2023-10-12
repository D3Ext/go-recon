package gorecon

import (
	"github.com/D3Ext/go-recon/core"
	"net/http"
	"time"
)

// function which aids users if they want to use
// a default client instance instead of creating a new one
func DefaultHttpClient() *http.Client {
	return core.DefaultHttpClient()
}

// create an http client with given timeout (in milliseconds),
// skip tls verify and some other useful settings
// don't follow redirects
// Example: client := CreateHttpClient(5000)
func CreateHttpClient(timeout int) *http.Client {
	return core.CreateHttpClient(timeout)
}

// this functions does the same as CreateHttpClient() but this one follows redirects
func CreateHttpClientFollowRedirects(timeout int) *http.Client {
	return core.CreateHttpClientFollowRedirects(timeout)
}

// return current time for later chaining
// with TimerDiff() to get elapsed time
func StartTimer() time.Time {
	return core.StartTimer()
}

// this function receives a time and
// returns difference between current time and given time
func TimerDiff(t1 time.Time) time.Duration {
	return core.TimerDiff(t1)
}
