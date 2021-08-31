package fio

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	defaultBaseURL = "https://www.fio.cz/"
	dateFormat     = "2006-01-02"
)

// NewClient returns new fio http client.
func NewClient(token string, client *http.Client) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)
	if client == nil {
		client = http.DefaultClient
	}

	c := &Client{
		BaseURL: baseURL,
		Token:   token,
		client:  client,
	}
	c.Transactions = &TransactionsService{client: c}
	return c
}

// Client is fio http api client.
type Client struct {
	client *http.Client

	Token        string
	BaseURL      *url.URL
	Transactions *TransactionsService
}

// NewRequest prepares new http request.
func (c *Client) newRequest(method string, urlStr string, body interface{}) (*http.Request, error) {
	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(body); err != nil {
			return nil, err
		}
	}
	return http.NewRequest(method, urlStr, buf)
}

// Get prepares new GET http request.
func (c *Client) get(urlStr string) (*http.Request, error) {
	return c.newRequest(http.MethodGet, urlStr, nil)
}

// Do performs http request and returns response.
func (c *Client) do(ctx context.Context, req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)
	resp, err := c.client.Do(req)
	if err != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		return nil, err
	}

	if err := c.checkResponse(resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) buildURL(resource string, segments ...string) string {
	var parts []string
	parts = append(parts, resource, c.Token)
	parts = append(parts, segments...)
	urlStr := strings.Join(parts, "/")
	ref, _ := url.Parse(urlStr)
	u := c.BaseURL.ResolveReference(ref)
	return u.String()
}

// ErrorResponse wraps http response errors.
type ErrorResponse struct {
	Response *http.Response
	Message  string
	Token    string
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d %v",
		r.Response.Request.Method, SanitizeURL(r.Token, r.Response.Request.URL),
		r.Response.StatusCode, r.Message)
}

func (c *Client) checkResponse(r *http.Response) error {
	if c := r.StatusCode; http.StatusOK <= c && c <= 299 {
		return nil
	}
	resp := &ErrorResponse{
		Response: r,
		Token:    c.Token,
	}
	defer r.Body.Close()

	// this seems to be the current response code mapping for fio API
	// 500 validation errors
	// 500 invalid token
	// 409 rate limit (one request per 30 seconds is allowed)
	// 404 resource not found
	// 400 invalid date format in url
	// 200 ok

	switch {
	// try to handle validation error
	case r.StatusCode == http.StatusInternalServerError && strings.Contains(r.Header.Get("Content-Type"), "text/xml"):
		var errResp xmlErrorResponse
		if err := xml.NewDecoder(r.Body).Decode(&errResp); err == nil {
			resp.Message = errResp.Result.Message
		}
	}

	return resp
}

// SanitizeURL redacts the token part of the URL.
func SanitizeURL(token string, u *url.URL) *url.URL {
	if token == "" {
		return u
	}

	redacted := strings.Replace(u.String(), token, "REDACTED", -1)
	redactedURL, _ := url.Parse(redacted)
	return redactedURL
}
