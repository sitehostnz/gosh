package srs

import (
	"context"
	"fmt"

	"github.com/google/go-querystring/query"
	"github.com/sitehostnz/gosh/pkg/models"
	"github.com/sitehostnz/gosh/pkg/net"
)

// EmailTemplate is a single registry-email template (renewal
// reminders, transfer confirmations, etc.).
//
// Field decoding matches the live response from
// list_email_templates: lowercase `name` is the human-readable
// label, `type` is the slug used internally, `subject` and
// `template` are the editable bodies. AvailableTags and
// RequiredTags list the {PLACEHOLDER} substitutions the registry
// recognises in this template.
type EmailTemplate struct {
	TemplateID    string   `json:"template_id"`
	ClientID      string   `json:"client_id"`
	Type          string   `json:"type"`
	Name          string   `json:"name"`
	Subject       string   `json:"subject"`
	Template      string   `json:"template"`
	DateAdded     string   `json:"date_added"`
	DateUpdated   string   `json:"date_updated"`
	AvailableTags []string `json:"available_tags"`
	RequiredTags  []string `json:"required_tags"`
	Customized    bool     `json:"customized"` //nolint:misspell // matches the upstream API JSON key
}

// ListEmailTemplatesResponse is the response from
// "srs/list_email_templates.json".
type ListEmailTemplatesResponse struct {
	Return []EmailTemplate `json:"return"`
	models.APIResponse
}

// ListEmailTemplates returns all email templates configured for
// the authenticated client via "srs/list_email_templates.json"
// (GET). The template body and subject are returned per template
// alongside its name.
//
// The public docs do not list parameters; only client_id (auto-
// injected by the SDK). Validate the response shape live before
// depending on field names.
func (s *Client) ListEmailTemplates(ctx context.Context) (response ListEmailTemplatesResponse, err error) {
	req, err := s.client.NewRequest("GET", "srs/list_email_templates.json", "")
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}

// GetEmailTemplateOptions identifies a single template to read.
// Template is the template's name as returned by
// ListEmailTemplates.
type GetEmailTemplateOptions struct {
	Template string `url:"template"`
}

// GetEmailTemplateResponse holds a single template body.
type GetEmailTemplateResponse struct {
	Return EmailTemplate `json:"return"`
	models.APIResponse
}

// GetEmailTemplate returns a single email template via
// "srs/get_email_templates.json" (GET). Note the docs spell the
// path with the trailing 's' on "templates" even for a single-
// template lookup.
//
// **Live finding (May 2026, gosh):** every plausible value derived
// from ListEmailTemplates output — the numeric `template_id`
// ("11239" or its "-13" prefixed form), the type slug
// ("AutoRenewReminder"), the lowercase variant, the human name
// ("Auto-Renew Reminder - 7 Days") — returns
//
//	200 Error: The specified template doesnt exist, or you dont
//	have access to it.
//
// (sic — API text omits the apostrophes). The lookup-key namespace
// for this endpoint appears disjoint from ListEmailTemplates'
// output. Wrapped here for completeness; consumers should expect
// API-level rejection until the expected input form is clarified.
//
// The wrapper's own SDK-level guard fires on empty input
// ("Template is required") before reaching the API — the API's
// own missing-parameter rejection ("The template name is missing.")
// isn't normally observable.
func (s *Client) GetEmailTemplate(ctx context.Context, opt GetEmailTemplateOptions) (response GetEmailTemplateResponse, err error) {
	if opt.Template == "" {
		return response, fmt.Errorf("srs.GetEmailTemplate: Template is required")
	}
	path, err := net.AddOptions("srs/get_email_templates.json", opt)
	if err != nil {
		return response, err
	}
	req, err := s.client.NewRequest("GET", path, "")
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}

// UpdateEmailTemplateOptions edits a single template.
//
// Template is the template name (required). EmailSubject and
// EmailTemplate are both optional — only set fields are sent. To
// reset a template to its default, the SiteHost API does not
// expose a dedicated "reset" verb; capture the defaults first via
// GetEmailTemplate before editing if a rollback is needed.
//
// **Account-wide impact:** these templates drive customer-facing
// renewal reminders, transfer notifications, etc. Test changes
// against a non-production account first.
type UpdateEmailTemplateOptions struct {
	Template      string `url:"template"`
	EmailSubject  string `url:"params[EmailSubject],omitempty"`
	EmailTemplate string `url:"params[EmailTemplate],omitempty"`
}

// UpdateEmailTemplate edits an email template via
// "srs/update_email_template.json" (POST).
func (s *Client) UpdateEmailTemplate(ctx context.Context, opt UpdateEmailTemplateOptions) (response models.APIResponse, err error) {
	if opt.Template == "" {
		return response, fmt.Errorf("srs.UpdateEmailTemplate: Template is required")
	}
	values, err := query.Values(opt)
	if err != nil {
		return response, err
	}
	req, err := s.client.NewRequest("POST", "srs/update_email_template.json", values.Encode())
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
