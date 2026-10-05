package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sitehostnz/gosh/pkg/api/dns"
	"github.com/sitehostnz/gosh/pkg/api/dns/template"
)

// stepTemplate points the zone at a different DNS template and puts it
// back.
//
// # What a template is for
//
// A template is a set of records applied to many zones at once, so
// that a hosting platform can change every customer's MX records in
// one place. Linking a zone to a template does not itself rewrite the
// zone; UpdateDomainDNS is what rebuilds a zone from its template, and
// it returns a scheduler job rather than completing inline.
//
// # The listing is not scoped to the account
//
// List returns SiteHost's shared templates alongside the account's
// own. Filter on ClientID when you want only yours, and do not read
// DomainCount on a shared row as an account figure — it is not scoped
// to the caller.
//
// Template id "0" is a real, usable template rather than a null id,
// which is worth knowing before treating 0 as "unset".
func stepTemplate(ctx context.Context, c clients, st *state) error {
	if !st.created {
		return fmt.Errorf("the template step changes the zone's linkage; run the journey so it acts on a zone this process created")
	}

	target, err := pickTemplate(ctx, c, st.originalTemplate)
	if err != nil {
		return err
	}
	if target == "" {
		log.Printf("  only one template is available; nothing to change to, skipping")
		return nil
	}

	if err := relink(ctx, c, st.cfg.zone, target); err != nil {
		return err
	}
	log.Printf("✓ %s now linked to template %s (was %s)", st.cfg.zone, target, st.originalTemplate)

	// Put it back, and assert that too — a restore that silently fails
	// would leave the zone on a template the journey chose.
	if err := relink(ctx, c, st.cfg.zone, st.originalTemplate); err != nil {
		return fmt.Errorf("restoring: %w", err)
	}
	log.Printf("✓ restored to template %s", st.originalTemplate)
	return nil
}

// pickTemplate reports the templates visible and chooses one that is
// not the zone's current template, so the assertion can actually fail.
func pickTemplate(ctx context.Context, c clients, current string) (string, error) {
	time.Sleep(throttle)
	templates, err := c.template.List(ctx)
	if err != nil {
		return "", fmt.Errorf("template.List: %w", err)
	}
	if len(templates.Return) == 0 {
		return "", fmt.Errorf("template.List returned nothing; the shared templates are always present")
	}

	var owned, shared int
	for _, t := range templates.Return {
		if t.ClientID == "0" {
			shared++
			continue
		}
		owned++
	}
	log.Printf("✓ %d template(s) visible: %d shared, %d belonging to this account",
		len(templates.Return), shared, owned)

	for _, t := range templates.Return {
		if t.TemplateID != current {
			return t.TemplateID, nil
		}
	}
	return "", nil
}

// relink points the zone at a template and checks it took effect.
func relink(ctx context.Context, c clients, zone, target string) error {
	time.Sleep(throttle)
	details, err := c.template.Get(ctx, template.GetRequest{TemplateID: target})
	if err != nil {
		return fmt.Errorf("template.Get %s: %w", target, err)
	}
	// get_template answers with a list holding one element, not the
	// object the name implies.
	if len(details.Return) != 1 {
		return fmt.Errorf("template.Get returned %d element(s), want 1", len(details.Return))
	}
	log.Printf("  template %s: nameserver %s, refresh %s",
		target, details.Return[0].Nameserver, details.Return[0].Refresh)

	time.Sleep(throttle)
	if _, err := c.template.UpdateDomain(ctx, template.UpdateDomainRequest{
		Domain: zone, TemplateID: target,
	}); err != nil {
		return fmt.Errorf("UpdateDomain: %w", err)
	}

	// Read the linkage back rather than trusting the status.
	linked, err := zoneTemplate(ctx, c, zone)
	if err != nil {
		return err
	}
	if linked != target {
		return fmt.Errorf("zone reports template %q after linking to %q; the change did not take effect", linked, target)
	}
	return nil
}

// zoneTemplate reads which template a zone is currently linked to.
func zoneTemplate(ctx context.Context, c clients, zone string) (string, error) {
	time.Sleep(throttle)
	found, err := c.dns.GetZone(ctx, dns.GetZoneRequest{DomainName: zone})
	if err != nil {
		return "", fmt.Errorf("GetZone: %w", err)
	}
	for _, z := range found.Return {
		if z.Name == zone {
			return z.TemplateID, nil
		}
	}
	return "", fmt.Errorf("zone %s is not in the search results", zone)
}
