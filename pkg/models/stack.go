package models

import (
	"encoding/json"

	"github.com/sitehostnz/gosh/pkg/shtypes"
)

type (
	// Container represents a container inside a stack.
	Container struct {
		Name        string      `json:"name"`
		ContainerID string      `json:"container_id"`
		State       string      `json:"state"`
		Size        string      `json:"size"`
		DateCreated string      `json:"date_created"`
		SslEnabled  bool        `json:"ssl_enabled"`
		IsMissing   interface{} `json:"is_missing"`
		Pending     interface{} `json:"pending"`

		// Image is the fully-qualified registry reference the
		// container runs, and ImageID the catalogue id behind it.
		//
		// Note ImageID arrives as a JSON number here while the
		// same value is a quoted string elsewhere in the API,
		// which is why it is not a plain string.
		Image   string              `json:"image"`
		ImageID shtypes.MaybeBigInt `json:"image_id"`

		// ImageDetails is deliberately left raw.
		//
		// Its keys are those of [StackImage], but two of them do
		// not agree with it. Labels is a JSON-encoded string here
		// and a real object in the image catalogue; versions is an
		// object of {id, labels, latest_version} here and a list
		// of versions there. Decoding it as a StackImage
		// therefore fails, and declaring a second near-identical
		// type would invite using the wrong one.
		//
		// Unmarshal it yourself if you need it, and check the
		// shape you get rather than assuming.
		ImageDetails json.RawMessage `json:"image_details"`

		// DockerSize is the image size in bytes, sent as a number
		// here and as a quoted string on stack image versions.
		DockerSize shtypes.MaybeBigInt `json:"docker_size"`

		// Backups and Monitored are real JSON booleans, unlike the
		// string flags used elsewhere on this API.
		Backups   bool `json:"backups"`
		Monitored bool `json:"monitored"`

		// DBSocket has only ever been observed as null. It is
		// typed loosely rather than guessed at; if you see a value
		// here, it is worth recording what it was.
		DBSocket interface{} `json:"db_socket"`

		DateAdded   string `json:"date_added"`
		DateUpdated string `json:"date_updated"`
	}

	// Stack represents a cloud stack and it's configuration.
	Stack struct {
		ClientID    string      `json:"client_id"`
		ServerID    string      `json:"server_id"`
		Server      string      `json:"server_name"`
		ServerLabel string      `json:"server_label"`
		Name        string      `json:"name"`
		Label       string      `json:"label"`
		DockerFile  string      `json:"docker_file"`
		IPAddress   string      `json:"ip_addr_server"`
		DateAdded   string      `json:"date_added"`
		DateUpdated string      `json:"date_updated"`
		Containers  []Container `json:"containers"`
		ServerOwner bool        `json:"server_owner"`
		Pending     interface{} `json:"pending"`
		IsMissing   interface{} `json:"is_missing"`
	}

	// EnvironmentVariable is a stack environment variable key-pair.
	EnvironmentVariable struct {
		Name    string `json:"name"`
		Content string `json:"content"`
	}
)
