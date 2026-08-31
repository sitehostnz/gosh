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
