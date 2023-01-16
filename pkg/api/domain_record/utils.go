package domain_record

import (
	"github.com/sitehostnz/gosh/pkg/models"
	"strings"
)

func adjustContent(domainRecord *models.DomainRecord) string {
	switch domainRecord.Type {
	// these need to have trailing spots... so lets fiddle them
	// not sure what other types do or don't need them
	case "NS", "CNAME":
		if !strings.HasSuffix(".", domainRecord.Content) {
			domainRecord.Content = domainRecord.Content + "."
		}
	}
	return domainRecord.Content
}
