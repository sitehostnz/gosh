package utils

import (
	"net/url"
	"reflect"
	"strings"

	"github.com/google/go-querystring/query"
)

// Encode generates a URL string in order.
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

// AddOptions adds the options to the URL.
func AddOptions(s string, opt interface{}) (string, error) {
	v := reflect.ValueOf(opt)

	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}

	origURL, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	origValues := origURL.Query()

	newValues, err := query.Values(opt)
	if err != nil {
		return s, err
	}

	for k, v := range newValues {
		origValues[k] = v
	}

	origURL.RawQuery = origValues.Encode()
	return origURL.String(), nil
}

// ConstructFqdn This is a simple helper to handle mapping of partial names and shortcuts to full names.
func ConstructFqdn(name, domain string) string {
	domain = strings.ToLower(domain)
	name = strings.ToLower(name)

	// let's put a dot at the end if we don't have one
	if !strings.HasSuffix(domain, ".") {
		domain += "."
	}

	// if we have an @ or a ., we can return the domain, as it's the apex
	// this is probably catering to something that only wizards would care about
	if name == "@" || name == "." {
		return domain
	}

	// if the name has the fqdn suffix at the end, we can return it, it's already a fqdn
	if strings.HasSuffix(name, domain) {
		return name
	}

	// otherwise, we need to append the suffix and the name together
	return strings.Join([]string{name, domain}, ".")
}

// DeconstructFqdn pulls off the suffix, the opposite of above and gives us an "@" for the apex.
func DeconstructFqdn(name, domain string) string {
	domain = strings.ToLower(domain)
	name = strings.ToLower(name)

	if !strings.HasSuffix(domain, ".") {
		domain += "."
	}

	if !strings.HasSuffix(name, ".") {
		name += "."
	}

	if name == domain {
		return "@"
	}

	return strings.TrimSuffix(name, "."+domain)
}
