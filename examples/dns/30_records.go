package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sitehostnz/gosh/pkg/api/dns"
	"github.com/sitehostnz/gosh/pkg/models"
)

// stepRecords walks a record through its whole life: add, read back,
// change, read back, remove, confirm gone.
//
// # Why each half is asserted separately
//
// Every one of these endpoints answers with a bare status. A caller
// that trusts the status has learned nothing about whether the zone
// changed — and the SDK has already shipped calls that reported
// success while a parameter was being dropped before it reached the
// wire. So each mutation is followed by a read that would fail if it
// had not taken effect.
//
// # The record name is fully qualified
//
// Name is "www.example.com", not "www". Sending the short form creates
// a record for "www.example.com.example.com", which the API accepts
// without complaint because it is a legal name.
func stepRecords(ctx context.Context, c clients, st *state) error {
	name := "sdk-test." + st.cfg.zone
	if err := addRecord(ctx, c, st, name); err != nil {
		return err
	}
	if err := updateRecord(ctx, c, st, name); err != nil {
		return err
	}
	return removeRecord(ctx, c, st)
}

// addRecord creates the record and checks it reads back as sent.
func addRecord(ctx context.Context, c clients, st *state, name string) error {
	const content = "203.0.113.10"

	time.Sleep(throttle)
	added, err := c.dns.AddRecord(ctx, dns.AddRecordRequest{
		Domain: st.cfg.zone, Type: "A", Name: name, Content: content,
	})
	if err != nil {
		return fmt.Errorf("AddRecord: %w", err)
	}
	st.recordID = added.Return.ID
	if st.recordID == "" {
		return fmt.Errorf("AddRecord reported success but returned no record id; nothing downstream can address it")
	}
	log.Printf("✓ added A %s -> %s (id %s)", name, content, st.recordID)

	got, err := findRecord(ctx, c, st.cfg.zone, st.recordID)
	if err != nil {
		return err
	}
	if got.Content != content {
		return fmt.Errorf("record content is %q, want %q — the add did not take effect", got.Content, content)
	}
	if got.Name != name {
		return fmt.Errorf("record name is %q, want %q — note names are fully qualified", got.Name, name)
	}
	log.Printf("✓ reads back with the content that was sent")
	return nil
}

// updateRecord replaces the record and checks the change is visible.
//
// Every field is required: this endpoint replaces the record rather
// than patching it, so omitting one blanks it.
func updateRecord(ctx context.Context, c clients, st *state, name string) error {
	const content = "203.0.113.20"

	time.Sleep(throttle)
	if _, err := c.dns.UpdateRecord(ctx, dns.UpdateRecordRequest{
		Domain: st.cfg.zone, RecordID: st.recordID,
		Type: "A", Name: name, Content: content,
	}); err != nil {
		return fmt.Errorf("UpdateRecord: %w", err)
	}

	got, err := findRecord(ctx, c, st.cfg.zone, st.recordID)
	if err != nil {
		return err
	}
	if got.Content != content {
		return fmt.Errorf("after update content is %q, want %q — the update reported success without taking effect", got.Content, content)
	}
	log.Printf("✓ updated to %s, and the change is visible", content)
	return nil
}

// removeRecord deletes the record and confirms it is gone.
//
// A delete that reports success and leaves the record behind is the
// case worth catching.
func removeRecord(ctx context.Context, c clients, st *state) error {
	time.Sleep(throttle)
	if _, err := c.dns.DeleteRecord(ctx, dns.DeleteRecordRequest{
		Domain: st.cfg.zone, RecordID: st.recordID,
	}); err != nil {
		return fmt.Errorf("DeleteRecord: %w", err)
	}

	time.Sleep(throttle)
	records, err := c.dns.ListRecords(ctx, dns.ListRecordsRequest{Domain: st.cfg.zone})
	if err != nil {
		return fmt.Errorf("ListRecords after delete: %w", err)
	}
	for _, r := range records.Return {
		if r.ID == st.recordID {
			return fmt.Errorf("record %s still present after DeleteRecord reported success", st.recordID)
		}
	}
	st.recordID = ""
	log.Printf("✓ deleted, and confirmed absent from the listing")
	return nil
}

// findRecord reads one record back from the zone listing.
//
// There is a GetRecord endpoint, but it searches by name and type
// rather than by id, so it cannot distinguish two records that differ
// only in content. The listing is the reliable way to check an id.
func findRecord(ctx context.Context, c clients, zone, id string) (models.DNSRecord, error) {
	time.Sleep(throttle)
	records, err := c.dns.ListRecords(ctx, dns.ListRecordsRequest{Domain: zone})
	if err != nil {
		return models.DNSRecord{}, fmt.Errorf("ListRecords: %w", err)
	}
	for _, r := range records.Return {
		if r.ID == id {
			return r, nil
		}
	}
	return models.DNSRecord{}, fmt.Errorf("record %s is not in the zone listing", id)
}
