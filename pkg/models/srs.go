package models

import "github.com/sitehostnz/gosh/pkg/shtypes"

type (
	Domain struct {
		ID     shtypes.MaybeBigInt `json:"domain_id"`
		Domain string              `json:"domain"`

		State   string            `json:"state"`
		Locked  shtypes.MaybeBool `json:"locked"`
		Private shtypes.MaybeBool `json:"private"`
		Pending shtypes.MaybeBool `json:"pending"`
		Premium shtypes.MaybeBool `json:"premium"`

		DateRegistered  string `json:"dateregistered"`
		DateModified    string `json:"datemodified"`
		DateBilledUntil string `json:"datebilleduntil"`
		DateCancelled   string `json:"datecancelled"`
		DateLocked      string `json:"datelocked"`

		AutorenewTerm          shtypes.MaybeBigInt `json:"autorenew_term"`
		AutorenewDaysRemaining shtypes.MaybeBigInt `json:"autorenew_days_remaining"`

		RegistrantName string `json:"registrant_name"`

		RegID   shtypes.MaybeBigInt `json:"reg_id"`
		RegName string              `json:"reg_name"`
		AdmID   shtypes.MaybeBigInt `json:"adm_id"`
		AdmName string              `json:"adm_name"`
		TecID   shtypes.MaybeBigInt `json:"tec_id"`
		TecName string              `json:"tec_name"`

		ClientID   shtypes.MaybeBigInt `json:"client_id"`
		ClientName string              `json:"client_name"`

		Api string `json:"api"`
	}

	DomainContact struct {
		ID             shtypes.MaybeBigInt `json:"contact_id"`
		Name           string              `json:"name"`
		RegistrantName string              `json:"registrant_name"`
		Email          string              `json:"email"`
		PhoneCountry   string              `json:"phone_cntry"`
		PhoneArea      string              `json:"phone_area"`
		PhoneLocal     string              `json:"phone_local"`
		DomainCount    shtypes.MaybeBigInt `json:"domain_count"`
	}
)
