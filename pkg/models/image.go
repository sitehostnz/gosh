package models

import "github.com/sitehostnz/gosh/pkg/utils"

type (

	// CloudImage represents an image in the /cloud/images.
	CloudImage struct {
		ID       string `json:"id"`
		ClientID string `json:"client_id"`
		Label    string `json:"label"`
		Code     string `json:"code"`
		Version  string `json:"version"`
		// this is a string in the api docs for the cloud image call...
		// but below, in the cloud stack image call, it's a struct...
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
	// StackImage represents an image in the /cloud/stack/images.
	StackImage struct {
		ID             string                 `json:"id"`
		ClientID       string                 `json:"client_id"`
		Label          string                 `json:"label"`
		Code           string                 `json:"code"`
		Version        string                 `json:"version"`
		Labels         map[string]interface{} `json:"labels"`
		Changelog      string                 `json:"changelog"`
		DateAdded      string                 `json:"date_added"`
		DateUpdated    string                 `json:"date_updated"`
		IsPublic       utils.MaybeBool        `json:"is_public"`
		IsMissing      utils.MaybeBool        `json:"is_missing"`
		ProjectID      string                 `json:"project_id"`
		RegistryID     string                 `json:"registry_id"`
		ForkedFrom     string                 `json:"forked_from"`
		Pending        string                 `json:"pending"`
		ClientName     string                 `json:"client_name"`
		ImageType      string                 `json:"image_type"`
		RegistryURL    string                 `json:"registry_url"`
		VersionCount   int                    `json:"version_count"`
		ContainerCount int                    `json:"container_count"`
		BuildStatus    string                 `json:"build_status"`
	}
)
