package bandwidth

import "github.com/sitehostnz/gosh/pkg/models"

type (
	ListIpAddressesResponse struct {
		models.APIResponse
		//Return struct {
		//	IPAddresses map[string]models.IPAddress `json:"subnets"`
		//	Subnet      map[string]models.IPAddress `json:"addresses"`
		//}
		Return map[string]models.IPAddress `json:"return"`
	}
)
