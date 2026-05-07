package srs

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// CompanyInfo describes the client's company profile.
	CompanyInfo struct {
		ClientID             int    `json:"client_id"`
		CompanyName          string `json:"companyname"`
		CompanyURL           string `json:"companyurl"`
		CompanyRenewURL      string `json:"companyrenewurl"`
		CompanyEmail         string `json:"companyemail"`
		CompanyEmailFrom     string `json:"companyemailfrom"`
		CompanyEmailFromName string `json:"companyemailfromname"`
		CompanySupportEmail  string `json:"companysupportemail"`
		CompanyPhone         string `json:"companyphone"`
		CompanyFax           string `json:"companyfax"`
		SendRenewedEmail     string `json:"send_renewed_email"`
		SendInvoice          string `json:"send_invoice"`
	}

	// GetCompanyInfoResponse represents the response from get_company_info.
	GetCompanyInfoResponse struct {
		Return CompanyInfo `json:"return"`
		models.APIResponse
	}
)
