package domain_record

type (
	RecordRequest struct {
		Id         string `json:"id"`
		RRType     string `json:"rr_type"`
		DomainName string `json:"name"`
	}

	ZoneRequest struct {
		DomainName string `json:"name"`
	}
)
