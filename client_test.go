package fio

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	// mux is the HTTP request multiplexer used with the test server.
	mux *http.ServeMux

	// client is the GitHub client being tested.
	client *Client

	// server is a test HTTP server used to provide mock API responses.
	server *httptest.Server

	// testingToken is token used across all tests.
	testingToken = "xxxx"
)

// setup sets up a test HTTP server along with a github.Client that is
// configured to talk to that test server. Tests should register handlers on
// mux which provide mock responses for the API method being tested.
func setup() {
	// test server
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	// github client configured to use test server
	client = NewClient(testingToken, nil)
	url, _ := url.Parse(server.URL)
	client.BaseURL = url
}

// teardown closes the test HTTP server.
func teardown() {
	server.Close()
}

var (
	sanitizeURLCases = []struct {
		token    string
		original string
		want     string
	}{
		{
			token:    "",
			original: "/dsa",
			want:     "/dsa",
		},
		{
			token:    "xx",
			original: "/dsa",
			want:     "/dsa",
		},
		{
			token:    "ds",
			original: "/dsa",
			want:     "/REDACTEDa",
		},
		{
			token:    "ds",
			original: "/dsads",
			want:     "/REDACTEDaREDACTED",
		},
	}
)

func TestSanitizeURL(t *testing.T) {
	for _, c := range sanitizeURLCases {
		t.Run(c.token+c.original+c.want, func(t *testing.T) {
			urlOrig := &url.URL{Path: c.original}
			urlGot := sanitizeURL(c.token, urlOrig)
			assert.Equal(t, c.want, urlGot.Path)
		})
	}
}
