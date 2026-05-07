package srs

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/go-querystring/query"
	"github.com/sitehostnz/gosh/pkg/models"
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

// ErrEmailTemplateUnsupported is returned by GetEmailTemplate to
// signal that the underlying API endpoint is not currently usable
// from this SDK. The lookup-key namespace expected by
// /srs/get_email_templates.json appears disjoint from anything
// surfaced by ListEmailTemplates (numeric template_id, type slug,
// human name — none accepted), and there's no documented form that
// satisfies it. Use ListEmailTemplates to read template bodies in
// the meantime.
//
// Match with errors.Is so the SDK can lift this to a real
// implementation later without a breaking change at the call site.
var ErrEmailTemplateUnsupported = errors.New(
	"srs.GetEmailTemplate: endpoint not currently usable — " +
		"every probed input form is rejected by the API with " +
		"\"The specified template doesnt exist, or you dont have " +
		"access to it.\" (sic). Use ListEmailTemplates to read " +
		"template bodies until the expected input form is clarified",
)

// GetEmailTemplate is the wrapper for "srs/get_email_templates.json"
// (GET) — note the trailing 's' on "templates" even for a single-
// template lookup.
//
// **Currently unusable.** Live probing (May 2026) exhausted every
// plausible value derived from ListEmailTemplates output — the
// numeric `template_id` ("11239" or its "-13" prefixed form), the
// type slug ("AutoRenewReminder"), the lowercase variant, the human
// name ("Auto-Renew Reminder - 7 Days") — and every one was
// rejected with
//
//	200 Error: The specified template doesnt exist, or you dont
//	have access to it.
//
// (sic — API text omits the apostrophes). The lookup-key namespace
// for this endpoint appears disjoint from anything ListEmailTemplates
// surfaces.
//
// To avoid relying on doc-comment-only warnings, the wrapper
// short-circuits and returns [ErrEmailTemplateUnsupported] without
// making the API call. Consumers can detect this with errors.Is.
// When the upstream input shape is clarified, the wrapper can be
// switched to making the real call without breaking call sites.
func (s *Client) GetEmailTemplate(_ context.Context, opt GetEmailTemplateOptions) (response GetEmailTemplateResponse, err error) {
	if opt.Template == "" {
		return response, fmt.Errorf("srs.GetEmailTemplate: Template is required")
	}
	return response, ErrEmailTemplateUnsupported
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
