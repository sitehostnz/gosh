package domain_record

import (
	"github.com/sitehostnz/gosh/pkg/models"
	"strings"
)

func NormaliseContent(domainRecord models.DomainRecord) string {
	switch domainRecord.Type {
	// these need to have trailing spots... so lets fiddle them
	// not sure what other types do or don't need them
	// but these are the main ones that we kinda car about, otherwise, we are gunna get weird stuff
	case "NS", "CNAME":
		if !strings.HasSuffix(domainRecord.Content, ".") {
			domainRecord.Content += "."
		}
	}

	return domainRecord.Content
}

func ConstructFqdn(name, domain string) string {

	if name == "@" {
		return domain
	}

	rn := strings.ToLower(name)
	rn = strings.TrimSuffix(rn, ".")

	if !strings.HasSuffix(rn, domain) {

		rn = strings.Join([]string{name, domain}, ".")
	}

	return rn
}

func DeconstructFqdn(name, domain string) string {
	if name == domain {
		return "@"
	}

	rn := strings.ToLower(name)
	rn = strings.TrimSuffix(rn, ".")
	rn = strings.TrimSuffix(rn, "."+domain)

	return rn
}
