package domain_record

import "github.com/sitehostnz/gosh/pkg/models"

type (
	//	ListResponse struct {
	//		//Server models.Server `json:"return"`
	//		Return struct {
	//			Domains *[]models.Domain `json:"data"`
	//		}
	//		models.APIResponse
	//	}
	ListResponse struct {
		DomainRecords *[]models.DomainRecord `json:"return"`
		Status        bool                   `json:"status"`
		Message       string                 `json:"msg"`
	}
)
