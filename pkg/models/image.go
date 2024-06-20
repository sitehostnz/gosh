package models

import (
	"bytes"
	"encoding/json"
)

type (
	// Image represents an image in the cloud images api.
	Image struct {
		ID       string `json:"id"`
		ClientID string `json:"client_id"`
		Label    string `json:"label"`
		Code     string `json:"code"`
		Labels   struct {
			NzSitehostImageVolumes struct {
				Source  any               `json:"source"`
				Hash    string            `json:"hash"`
				Volumes map[string]Volume `json:"volumes"`
			} `json:"nz.sitehost.image.volumes"`
			NzSitehostImageProvider string   `json:"nz.sitehost.image.provider"`
			NzSitehostImageType     string   `json:"nz.sitehost.image.type"`
			NzSitehostImageLabel    string   `json:"nz.sitehost.image.label"`
			NzSitehostImagePorts    PortList `json:"nz.sitehost.image.ports"`
		} `json:"labels"`
		Changelog   string `json:"changelog"`
		DateAdded   string `json:"date_added"`
		DateUpdated string `json:"date_updated"`
		IsPublic    string `json:"is_public"`
		IsMissing   string `json:"is_missing"`
		ProjectID   string `json:"project_id"`
		RegistryID  string `json:"registry_id"`
		ForkedFrom  any    `json:"forked_from"`
		Pending     any    `json:"pending"`
		Versions    []struct {
			ID          string `json:"id"`
			ClientID    string `json:"client_id"`
			ImageID     string `json:"image_id"`
			Version     string `json:"version"`
			Labels      string `json:"labels"`
			DateAdded   string `json:"date_added"`
			DateUpdated string `json:"date_updated"`
			IsMissing   string `json:"is_missing"`
			ForceConfig string `json:"force_config"`
			BuildID     string `json:"build_id"`
			BuildStatus string `json:"build_status"`
			Pending     string `json:"pending"`
		} `json:"versions"`
		RegistryPath string `json:"registry_path,omitempty"`
	}

	// PortList represents a port list in the cloud images api.
	PortList map[string]Port

	// Port represents a port in the cloud images api.
	Port struct {
		Protocol string `json:"protocol"`
		Publish  bool   `json:"publish"`
		Exposed  bool   `json:"exposed"`
	}

	// Volume represents a volume in the cloud images api.
	Volume struct {
		Dest string `json:"dest"`
		Mode string `json:"mode"`
	}
)

// UnmarshalJSON implements the json.Unmarshaler interface to fix the port list format.
func (p *PortList) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("[]")) {
		*p = make(map[string]Port)
		return nil
	}
	var m map[string]Port
	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}
	*p = m
	return nil
}
