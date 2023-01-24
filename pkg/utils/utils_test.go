package utils

import (
	"net/url"
	"testing"
)

func TestEncode(t *testing.T) {
	t.Parallel()
	v := url.Values{}
	v.Add("key1", "value1")
	v.Add("key2", "value2")
	v.Add("key3", "value3")
	keys := []string{"key1", "key2"}
	expected := "key1=value1&key2=value2"
	result := Encode(v, keys)
	if result != expected {
		t.Errorf("Expected %s but got %s", expected, result)
	}
}
