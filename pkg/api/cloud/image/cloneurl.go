package image

import "fmt"

// CloneURL returns the SSH git clone URL for a custom image's
// backing GitLab repository. The format is
//
//	git@<host>:g_<client_id>/<image_code>.git
//
// host defaults to "gitlab-clients.sitehost.co.nz" but can be
// overridden via api.SetCustomImageGitHost when constructing the
// underlying api.Client. client_id is the consumer's account ID
// (already on the api.Client). image_code is the image's Code as
// returned by Get / list_all.
//
// **Why this is a helper rather than an API field:** the GitLab
// repo URL isn't returned by any /cloud/image endpoint — only the
// Docker registry URL is (registry_url). Consumers were expected to
// copy the repo URL from the Control Panel UI; this helper closes
// that gap so SDK consumers can clone without leaving the Go
// program.
func (s *Client) CloneURL(imageCode string) string {
	return fmt.Sprintf("git@%s:g_%s/%s.git",
		s.client.CustomImageGitHost,
		s.client.ClientID,
		imageCode)
}
