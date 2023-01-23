package record

import (
	"strings"

	"github.com/sitehostnz/gosh/pkg/models"
)

func adjustContent(domainRecord *models.DNSRecord) string {
	switch domainRecord.Type {
	// these need to have trailing spots... so lets fiddle them
	// not sure what other types do or don't need them
	case "NS", "CNAME":
		if !strings.HasSuffix(".", domainRecord.Content) {
			domainRecord.Content += "."
		}
	}
	return domainRecord.Content
}
