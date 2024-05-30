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
	defaultBaseURL = "https://fioapi.fio.cz/"
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

func (c *Client) newGetRequest(ctx context.Context, urlStr string) (*http.Request, error) {
	return c.newRequest(ctx, http.MethodGet, urlStr, nil)
}

func (c *Client) newRequest(ctx context.Context, method string, urlStr string, body interface{}) (*http.Request, error) {
	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(body); err != nil {
			return nil, err
		}
	}
	return http.NewRequestWithContext(ctx, method, urlStr, buf)
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
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

	// try to handle validation error
	if r.StatusCode == http.StatusInternalServerError && strings.Contains(r.Header.Get("Content-Type"), "text/xml") {
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

	redacted := strings.ReplaceAll(u.String(), token, "REDACTED")
	redactedURL, _ := url.Parse(redacted)
	return redactedURL
}
