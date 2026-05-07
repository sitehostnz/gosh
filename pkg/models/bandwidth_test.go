package models

import (
	"encoding/json"
	"testing"
)

func TestIPAddress_UnmarshalJSON_IPv4Untouched(t *testing.T) {
	t.Parallel()
	in := `{"ip_addr":"203.0.113.10/32","netmask":"255.255.255.0","prefix":"32","reserved":"0","rdns":"","addr_family":"4","date_allocated":""}`
	var got IPAddress
	if err := json.Unmarshal([]byte(in), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.IP != "203.0.113.10/32" {
		t.Errorf("IP = %q, want untouched 203.0.113.10/32", got.IP)
	}
}

func TestIPAddress_UnmarshalJSON_IPv6CanonicalUntouched(t *testing.T) {
	t.Parallel()
	in := `{"ip_addr":"2403:7000:8000:300::ce/128","addr_family":"6"}`
	var got IPAddress
	if err := json.Unmarshal([]byte(in), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.IP != "2403:7000:8000:300::ce/128" {
		t.Errorf("IP = %q, want untouched canonical IPv6", got.IP)
	}
}

func TestIPAddress_UnmarshalJSON_IPv6MangledShortForm(t *testing.T) {
	t.Parallel()
	// '::' → '..' shorthand mangling, the form /bandwidth/get_ip_list.json emits.
	in := `{"ip_addr":"2403.7000.8000.300..ce/128","addr_family":"6"}`
	var got IPAddress
	if err := json.Unmarshal([]byte(in), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.IP != "2403:7000:8000:300::ce/128" {
		t.Errorf("IP = %q, want 2403:7000:8000:300::ce/128", got.IP)
	}
}

func TestIPAddress_UnmarshalJSON_IPv6MangledLongForm(t *testing.T) {
	t.Parallel()
	// IPv6 with all 8 hextets, no '::' shorthand — only single dots.
	in := `{"ip_addr":"2403.7000.8000.c00.0.0.0.9b/128","addr_family":"6"}`
	var got IPAddress
	if err := json.Unmarshal([]byte(in), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.IP != "2403:7000:8000:c00:0:0:0:9b/128" {
		t.Errorf("IP = %q, want 2403:7000:8000:c00:0:0:0:9b/128", got.IP)
	}
}

func TestIPAddress_UnmarshalJSON_NoPrefix(t *testing.T) {
	t.Parallel()
	in := `{"ip_addr":"2403.7000.8000.300..ce","addr_family":"6"}`
	var got IPAddress
	if err := json.Unmarshal([]byte(in), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.IP != "2403:7000:8000:300::ce" {
		t.Errorf("IP = %q, want 2403:7000:8000:300::ce", got.IP)
	}
}

func TestIPAddress_UnmarshalJSON_EmptyIP(t *testing.T) {
	t.Parallel()
	in := `{"ip_addr":"","addr_family":"6"}`
	var got IPAddress
	if err := json.Unmarshal([]byte(in), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.IP != "" {
		t.Errorf("IP = %q, want empty", got.IP)
	}
}

func TestIPAddress_UnmarshalJSON_FamilyMissingPassesThrough(t *testing.T) {
	t.Parallel()
	// Without addr_family="6" we don't know it's IPv6, so don't transform.
	// IPv4 has dots and we must not mangle it.
	in := `{"ip_addr":"203.0.113.10/32"}`
	var got IPAddress
	if err := json.Unmarshal([]byte(in), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.IP != "203.0.113.10/32" {
		t.Errorf("IP = %q, want untouched 203.0.113.10/32 (family unknown)", got.IP)
	}
}
