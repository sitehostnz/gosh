package net

import "strings"

// ConstructFqdn is a simple helper to handle mapping of partial names and shortcuts to full names.
func ConstructFqdn(name, domain string) string {
	domain = strings.ToLower(domain)
	name = strings.ToLower(name)

	if domain == "" {
		return domain
	}

	// let's put a dot at the end if we don't have one
	if !strings.HasSuffix(domain, ".") {
		domain += "."
	}

	// if we have an @ or a ., we can return the domain, as it's the apex
	// this is probably catering to something that only wizards would care about
	if name == "@" || name == "." || name == "" {
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
