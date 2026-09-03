package models

import (
	"github.com/sitehostnz/gosh/pkg/shtypes"
)

type (

	// CloudImage represents an image in the /cloud/image.
	CloudImage struct {
		ID       string `json:"id"`
		ClientID string `json:"client_id"`
		Label    string `json:"label"`
		Code     string `json:"code"`
		Version  string `json:"version"`
		// this is a string in the api docs for the cloud image call...
		// but below, in the cloud stack image call, it's a struct...
		Labels         string            `json:"labels"`
		Changelog      string            `json:"changelog"`
		DateAdded      string            `json:"date_added"`
		DateUpdated    string            `json:"date_updated"`
		IsPublic       shtypes.MaybeBool `json:"is_public"`
		IsMissing      shtypes.MaybeBool `json:"is_missing"`
		ProjectID      string            `json:"project_id"`
		RegistryID     string            `json:"registry_id"`
		ForkedFrom     string            `json:"forked_from"`
		Pending        string            `json:"pending"`
		ClientName     string            `json:"client_name"`
		ImageType      string            `json:"image_type"`
		RegistryURL    string            `json:"registry_url"`
		VersionCount   int               `json:"version_count"`
		ContainerCount int               `json:"container_count"`
		BuildStatus    string            `json:"build_status"`
	}

	// StackImage represents an image in the /cloud/stack/image.
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
		IsPublic       shtypes.MaybeBool      `json:"is_public"`
		IsMissing      shtypes.MaybeBool      `json:"is_missing"`
		ProjectID      string                 `json:"project_id"`
		RegistryID     string                 `json:"registry_id"`
		ForkedFrom     string                 `json:"forked_from"`
		Pending        interface{}            `json:"pending"`
		ClientName     string                 `json:"client_name"`
		ImageType      string                 `json:"image_type"`
		RegistryURL    string                 `json:"registry_url"`
		VersionCount   int                    `json:"version_count"`
		ContainerCount int                    `json:"container_count"`
		BuildStatus    string                 `json:"build_status"`
		Versions       []StackImageVersion    `json:"versions"`

		// RegistryPath is the client-scoped registry prefix a
		// custom image is pushed to. Public catalogue images do
		// not carry it, so an empty value is normal.
		RegistryPath string `json:"registry_path"`
	}

	// StackImageVersion stores the version information for a StackImage.
	StackImageVersion struct {
		ID          string      `json:"id"`
		ClientID    string      `json:"client_id"`
		ImageIDs    string      `json:"image_id"`
		Version     string      `json:"version"`
		Labels      string      `json:"labels"`
		DateAdded   string      `json:"date_added"`
		DateUpdated string      `json:"date_updated"`
		IsMissing   string      `json:"is_missing"`
		ForceConfig string      `json:"force_config"`
		BuildID     string      `json:"build_id"`
		BuildStatus string      `json:"build_status"`
		Pending     interface{} `json:"pending"`

		// DockerSize is the version's size in bytes, sent quoted
		// here and as a bare number on [Container.DockerSize].
		DockerSize shtypes.MaybeBigInt `json:"docker_size"`

		// Public arrives as "0" or "1". Note the neighbouring
		// [StackImage.IsPublic] spells the same idea is_public.
		Public shtypes.MaybeBool `json:"public"`
	}

	// Port is for storing the exposed port information for a container.
	Port struct {
		Exposed  bool
		Publish  bool
		Protocol string
	}

	// Volume is for storing the volume information for a container.
	Volume struct {
		Hash   string `json:"hash"`
		Source struct {
			Filename string `json:"filename"`
			URL      string `json:"url"`
		} `json:"source"`
		Type string `json:"type"`

		Volumes struct {
			Application struct {
				Dest string `json:"dest"`
				GID  int    `json:"gid"`
				Mode string `json:"mode"`
				UID  int    `json:"uid"`
			} `json:"application"`
		} `json:"volumes"`
	}
)
