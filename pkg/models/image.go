package models

import "github.com/sitehostnz/gosh/pkg/utils"

type (
	// Image represents an image in the cloud images api.
	Image struct {
		ID             string          `json:"id"`
		ClientID       string          `json:"client_id"`
		Label          string          `json:"label"`
		Code           string          `json:"code"`
		Version        string          `json:"version"`
		Labels         string          `json:"labels"`
		Changelog      string          `json:"changelog"`
		DateAdded      string          `json:"date_added"`
		DateUpdated    string          `json:"date_updated"`
		IsPublic       utils.MaybeBool `json:"is_public"`
		IsMissing      utils.MaybeBool `json:"is_missing"`
		ProjectID      string          `json:"project_id"`
		RegistryID     string          `json:"registry_id"`
		ForkedFrom     string          `json:"forked_from"`
		Pending        string          `json:"pending"`
		ClientName     string          `json:"client_name"`
		ImageType      string          `json:"image_type"`
		RegistryURL    string          `json:"registry_url"`
		VersionCount   int             `json:"version_count"`
		ContainerCount int             `json:"container_count"`
		BuildStatus    string          `json:"build_status"`
	}
)
