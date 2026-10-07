package service

import "testing"

func TestNormalizeEmail(t *testing.T) {
	got, err := NormalizeEmail(" Admin@Example.ORG ")
	if err != nil || got != "admin@example.org" {
		t.Fatalf("NormalizeEmail() = %q, %v", got, err)
	}
	for _, input := range []string{"", "bad", "Name <user@example.org>"} {
		if _, err := NormalizeEmail(input); err == nil {
			t.Errorf("NormalizeEmail(%q) should be rejected", input)
		}
	}
}
