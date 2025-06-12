package net

import "testing"

const exampleDotCom = "example.com."

func TestConstructFqdnWithAtSymbolAndName(t *testing.T) {
	t.Parallel()

	result := ConstructFqdn("@", "example.com")

	if result != exampleDotCom {
		t.Errorf("Expected %s but got %s", exampleDotCom, result)
	}
}

func TestConstructFqdnWithDotSymbolAndName(t *testing.T) {
	t.Parallel()
	result := ConstructFqdn(".", "example.com")

	if result != exampleDotCom {
		t.Errorf("Expected %s but got %s", exampleDotCom, result)
	}
}

func TestConstructFqdnWithAtSymbolAndNameAndTrailingDot(t *testing.T) {
	t.Parallel()
	result := ConstructFqdn("@", exampleDotCom)

	if result != exampleDotCom {
		t.Errorf("Expected %s but got %s", exampleDotCom, result)
	}
}

func TestConstructFqdnWithSubDomain(t *testing.T) {
	t.Parallel()
	expected := "www.example.com."
	result := ConstructFqdn("www", "example.com")

	if result != expected {
		t.Errorf("Expected %s but got %s", expected, result)
	}
}

func TestConstructFqdnWithFdqn(t *testing.T) {
	t.Parallel()
	expected := exampleDotCom
	result := ConstructFqdn(exampleDotCom, "example.com")

	if result != expected {
		t.Errorf("Expected %s but got %s", expected, result)
	}
}

func TestConstructFqdnWithFdqnAndSubDomain(t *testing.T) {
	t.Parallel()
	expected := "www.example.com."
	result := ConstructFqdn("www.example.com.", "example.com")

	if result != expected {
		t.Errorf("Expected %s but got %s", expected, result)
	}
}

func TestConstructFqdnWithNotFdqnAndSubDomain(t *testing.T) {
	t.Parallel()
	expected := "www.example.com.example.com."
	result := ConstructFqdn("www.example.com", "example.com")

	if result != expected {
		t.Errorf("Expected %s but got %s", expected, result)
	}
}

func TestDeconstructFqdnApex(t *testing.T) {
	t.Parallel()
	expected := "@"
	result := DeconstructFqdn(exampleDotCom, exampleDotCom)

	if result != expected {
		t.Errorf("Expected %s but got %s", expected, result)
	}
}

func TestDeconstructFqdnSubDomain(t *testing.T) {
	t.Parallel()
	expected := "www"
	result := DeconstructFqdn("www.example.com", "example.com")

	if result != expected {
		t.Errorf("Expected %s but got %s", expected, result)
	}
}
