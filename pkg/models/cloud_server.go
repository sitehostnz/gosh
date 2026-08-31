package models

import "github.com/sitehostnz/gosh/pkg/shtypes"

type (
	// CloudServer is a view of a server that can run stacks on. This server also exists under the server API.
	CloudServer struct {
		ID                  string   `json:"id"`
		ClientID            string   `json:"client_id"`
		Name                string   `json:"name"`
		Label               string   `json:"label"`
		State               string   `json:"state"`
		PrimaryIP           string   `json:"primary_ip"`
		IPAddrProxy         string   `json:"ip_addr_proxy"`
		ProductID           string   `json:"product_id"`
		Owner               bool     `json:"owner"`
		ImagesUsed          []string `json:"images_used"`
		ImagesRemaining     int      `json:"images_remaining"`
		ContainersRemaining int      `json:"containers_remaining"`

		// Created and DateUpdated are "YYYY-MM-DD HH:MM:SS" in NZ
		// time, with no zone marker. They were being dropped
		// silently until a recorded response was compared against
		// this type.
		Created     string `json:"created"`
		DateUpdated string `json:"date_updated"`

		// Managed arrives as the string "0" or "1", not a bool.
		// It is what decides whether the update-window endpoints
		// apply: on an unmanaged server they answer "This server
		// is not managed by SiteHost."
		Managed shtypes.MaybeBool `json:"managed"`
	}
)
