package gosh

import (
	"net/url"
	"strings"
)

// Encode generate a URL string in order
func Encode(v url.Values, keys []string) string {
	if v == nil {
		return ""
	}
	var buf strings.Builder

	for _, k := range keys {
		vs := v[k]
		keyEscaped := url.QueryEscape(k)
		for _, v := range vs {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(keyEscaped)
			buf.WriteByte('=')
			buf.WriteString(url.QueryEscape(v))
		}
	}
	return buf.String()
}
