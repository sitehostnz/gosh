package domain_record

type (
	GetRequest struct {
		Id         string `json:"id"`
		RRType     string `json:"rr_type"`
		DomainName string `json:"name"`
	}

	ListRequest struct {
		DomainName string `json:"name"`
	}
)
