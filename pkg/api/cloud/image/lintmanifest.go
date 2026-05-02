package image

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// LintError is returned by LintManifest when a manifest.yml fails
// schema validation. Errors holds one entry per problem found.
type LintError struct {
	Errors []string
}

// Error joins the lint errors into a single message.
func (e *LintError) Error() string {
	if len(e.Errors) == 1 {
		return "manifest lint: " + e.Errors[0]
	}
	return fmt.Sprintf("manifest lint: %d problems: %v", len(e.Errors), e.Errors)
}

// minimalManifest mirrors the required-field shape of a custom
// image's manifest.yml. We don't model the full schema (ports,
// volumes, env_file etc.) because the API doesn't enforce those
// strictly and consumers may use platform-specific extensions.
type minimalManifest struct {
	Version int `yaml:"version"`
	Image   struct {
		Label    string `yaml:"label"`
		Type     string `yaml:"type"`
		Provider string `yaml:"provider"`
	} `yaml:"image"`
}

// LintManifest validates the bytes of a custom-image manifest.yml
// against the documented minimum schema:
//
//   - top-level "version" key (must be 1)
//   - "image.label" non-empty
//   - "image.type" non-empty (typically "www" or a service name)
//   - "image.provider" non-empty (the customer/account label)
//
// Local lint is intentionally narrow: catch the "did you forget
// a required field" class of mistakes that otherwise only surface
// hours later in a failed CI build trace. The platform's CI
// remains the source of truth for the full schema.
func LintManifest(data []byte) error {
	var m minimalManifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return &LintError{Errors: []string{fmt.Sprintf("YAML parse: %v", err)}}
	}

	var errs []string
	if m.Version != 1 {
		errs = append(errs, fmt.Sprintf("version must be 1, got %d", m.Version))
	}
	if m.Image.Label == "" {
		errs = append(errs, `image.label is required`)
	}
	if m.Image.Type == "" {
		errs = append(errs, `image.type is required (e.g. "www")`)
	}
	if m.Image.Provider == "" {
		errs = append(errs, `image.provider is required`)
	}

	if len(errs) > 0 {
		return &LintError{Errors: errs}
	}
	return nil
}
