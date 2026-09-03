package recorder

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestScrub_PreservesShape is the property the whole idea rests on: a
// scrubbed fixture must decode into the same Go value shape as the
// recording it came from, or it proves nothing about the API.
func TestScrub_PreservesShape(t *testing.T) {
	t.Parallel()
	in := `{
		"status": true,
		"msg": "Successful",
		"return": {
			"total_items": 7,
			"data": [
				{"id": "45342", "db_name": "customerdb", "size": "10125312", "pending": null, "server_owner": true},
				{"id": "45343", "db_name": "otherdb", "size": null, "pending": null, "server_owner": false}
			]
		}
	}`

	out, err := Scrub([]byte(in))
	if err != nil {
		t.Fatalf("Scrub: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("scrubbed output does not decode: %v", err)
	}

	// The envelope's types are load-bearing for every response type.
	if _, ok := got["status"].(bool); !ok {
		t.Errorf("status = %T, want bool", got["status"])
	}
	ret, ok := got["return"].(map[string]any)
	if !ok {
		t.Fatalf("return = %T, want object", got["return"])
	}
	if _, ok := ret["total_items"].(float64); !ok {
		t.Errorf("total_items = %T, want a number — a count must not become a string", ret["total_items"])
	}

	rows, ok := ret["data"].([]any)
	if !ok || len(rows) != 2 {
		t.Fatalf("data = %T len %d, want 2 rows", ret["data"], len(rows))
	}
	row, ok := rows[0].(map[string]any)
	if !ok {
		t.Fatalf("rows[0] = %T, want object", rows[0])
	}

	// A quoted integer must stay quoted. This API sends ids as strings,
	// and a fixture that "helpfully" unquoted them would hide a decode
	// bug rather than catch one.
	if s, ok := row["id"].(string); !ok || s != "1" {
		t.Errorf("id = %#v, want the string \"1\"", row["id"])
	}
	if _, ok := row["server_owner"].(bool); !ok {
		t.Errorf("server_owner = %T, want bool", row["server_owner"])
	}

	// null must survive as null. Null-versus-absent is the distinction
	// hand-written fixtures get wrong most often.
	if v, present := row["pending"]; !present || v != nil {
		t.Errorf("pending = %#v present=%t, want an explicit null", v, present)
	}
	// And absent must stay absent.
	if _, present := row["is_missing"]; present {
		t.Error("is_missing appeared in the output; scrubbing must not invent keys")
	}
	// A null-valued string field stays null, not "".
	second, ok := rows[1].(map[string]any)
	if !ok {
		t.Fatalf("rows[1] = %T, want object", rows[1])
	}
	if v, present := second["size"]; !present || v != nil {
		t.Errorf("size = %#v, want null preserved", v)
	}
}

// TestScrub_RemovesEveryValue checks the safety half.
func TestScrub_RemovesEveryValue(t *testing.T) {
	t.Parallel()
	in := `{
		"return": {"data": [{
			"username": "acmecorp",
			"home_dir": "/data/docker0/ssh/acmecorp/login",
			"server_ip": "192.0.2.9",
			"ssh_keys": [{"content": "ssh-rsa AAAAB3NzaC1yc2EAAAA acme@example"}],
			"label": "acme production"
		}]}
	}`

	out, err := Scrub([]byte(in))
	if err != nil {
		t.Fatalf("Scrub: %v", err)
	}
	for _, leak := range []string{"acmecorp", "acme", "192.0.2.9", "AAAAB3NzaC1yc2E", "production"} {
		if strings.Contains(string(out), leak) {
			t.Errorf("scrubbed output still contains %q", leak)
		}
	}
}

// TestScrub_CapsArrays keeps fixtures reviewable. An 88-element image
// catalogue proves nothing an 2-element one does not.
func TestScrub_CapsArrays(t *testing.T) {
	t.Parallel()
	in := `{"return": [{"a":"1"},{"a":"2"},{"a":"3"},{"a":"4"},{"a":"5"}]}`
	out, err := Scrub([]byte(in))
	if err != nil {
		t.Fatalf("Scrub: %v", err)
	}
	var got struct {
		Return []map[string]string `json:"return"`
	}
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got.Return) != maxArrayElements {
		t.Errorf("len = %d, want %d", len(got.Return), maxArrayElements)
	}
}

// TestScrub_RejectsNonJSON checks a truncated or HTML body is an error
// rather than a fixture that silently says nothing.
func TestScrub_RejectsNonJSON(t *testing.T) {
	t.Parallel()
	if _, err := Scrub([]byte("<html>502 Bad Gateway</html>")); err == nil {
		t.Fatal("expected an error for a non-JSON body")
	}
}

// TestScrub_RemovesDataShapedKeys covers the half the value rules miss.
//
// Keys were passed through untouched, which is right for a struct and
// wrong for the responses this API keys by customer data:
// redirect.ListRedirects returns map[domain]map[sourceURL]Rule, and
// server.ListAllocatedIPs is keyed by the address itself — so scrubbing
// replaced IPAddr and left the address in the key above it.
func TestScrub_RemovesDataShapedKeys(t *testing.T) {
	t.Parallel()

	in := `{"return":{
		"acme-widgets-prod.co.nz":{"disk1":true},
		"jsmith@example.org":1,
		"203.0.113.9":{"ip_addr":"203.0.113.9"}
	}}`
	out, err := Scrub([]byte(in))
	if err != nil {
		t.Fatalf("Scrub: %v", err)
	}

	for _, leak := range []string{"acme-widgets-prod", "jsmith@example.org", "203.0.113.9"} {
		if strings.Contains(string(out), leak) {
			t.Errorf("scrubbed output still contains the key %q", leak)
		}
	}

	var got map[string]map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// All three entries must survive. Two data keys scrubbing to the
	// same placeholder would collide and collapse the map, losing the
	// entry count — which is part of the shape a fixture records.
	if len(got["return"]) != 3 {
		t.Errorf("return has %d entries, want 3 — scrubbed keys must not collide", len(got["return"]))
	}

	// A schema key is not data and must survive, or the fixture stops
	// describing the response.
	var kept bool
	for _, v := range got["return"] {
		if m, ok := v.(map[string]any); ok {
			if _, ok := m["disk1"]; ok {
				kept = true
			}
		}
	}
	if !kept {
		t.Error("the schema key disk1 was scrubbed; only data-shaped keys should be")
	}
}

// TestScrub_KeepsOnlyTheEnvelopeMessage pins the one value exception.
//
// The message is API-authored and the SDK branches on it, so it stays.
// Everything else goes, whatever it looks like — the first version of
// this exception tested the value's shape from inside scrubString,
// which every string leaf passes through, so a customer's label
// opening "The " or a note opening "This " survived verbatim. Those
// are ordinary openings for both.
func TestScrub_KeepsOnlyTheEnvelopeMessage(t *testing.T) {
	t.Parallel()

	in := `{"msg":"Successful","return":{
		"label":"The Big Client Ltd",
		"note":"This is my prod box",
		"d":"no idea",
		"e":"not for sharing",
		"f":"Please specify a name",
		"g":"Error: mine",
		"msg":"The nested one is not the envelope"
	}}`
	out, err := Scrub([]byte(in))
	if err != nil {
		t.Fatalf("Scrub: %v", err)
	}

	for _, leak := range []string{
		"Big Client", "prod box", "no idea", "not for sharing",
		"Please specify a name", "Error: mine", "nested one",
	} {
		if strings.Contains(string(out), leak) {
			t.Errorf("scrubbed output still contains %q — only the top-level msg is kept", leak)
		}
	}

	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got["msg"] != "Successful" {
		t.Errorf("msg = %v, want it kept — the SDK's control flow matches on it", got["msg"])
	}
}

// TestScrub_KeepsReverseDNSAndNumericKeys is the must-survive side of
// the key rule.
//
// A reverse-DNS namespace and a domain are the same shape, and telling
// them apart is what the first version got wrong: it rewrote
// nz.sitehost.image.* to scrubbed-N and the ports map's "80" to "1-1",
// erasing the only record in the repository of how those labels nest.
// The earlier test picked "disk1" as its schema key, which no pattern
// matched, so it reinforced the blind spot rather than catching it.
func TestScrub_KeepsReverseDNSAndNumericKeys(t *testing.T) {
	t.Parallel()

	in := `{"labels":{
		"nz.sitehost.image.ports":{"80":{"exposed":true},"443":{"exposed":false}},
		"nz.sitehost.image.volumes":{"source":"https://example.test/x"},
		"acme-widgets-prod.co.nz":{"a":1}
	}}`
	out, err := Scrub([]byte(in))
	if err != nil {
		t.Fatalf("Scrub: %v", err)
	}
	s := string(out)

	for _, keep := range []string{"nz.sitehost.image.ports", "nz.sitehost.image.volumes", `"80"`, `"443"`} {
		if !strings.Contains(s, keep) {
			t.Errorf("scrubbed output lost the schema key %s", keep)
		}
	}
	// And a real customer domain is still removed.
	if strings.Contains(s, "acme-widgets-prod") {
		t.Error("a customer domain survived as a key")
	}
}
