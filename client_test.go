package fio

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

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
			urlGot := SanitizeURL(c.token, urlOrig)
			require.Equal(t, c.want, urlGot.Path)
		})
	}
}
