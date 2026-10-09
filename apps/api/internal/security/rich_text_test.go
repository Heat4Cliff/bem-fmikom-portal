package security

import (
	"strings"
	"testing"
)

func TestSanitizeRichTextKeepsAllowedMarkupAndRemovesExecutableContent(t *testing.T) {
	got, err := SanitizeRichText(`<p onclick="run()">Halo <strong>BEM</strong><script>alert(1)</script><a href="javascript:alert(1)" onmouseover="run()">tautan</a><iframe src="https://evil.example"></iframe></p>`)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"onclick", "onmouseover", "javascript:", "<script", "<iframe", "alert(1)"} {
		if strings.Contains(strings.ToLower(got), forbidden) {
			t.Fatalf("sanitized content contains %q: %s", forbidden, got)
		}
	}
	if !strings.Contains(got, "<p>Halo <strong>BEM</strong>") || !strings.Contains(got, ">tautan</a>") {
		t.Fatalf("safe text/formatting was not preserved: %s", got)
	}
}

func TestSanitizeRichTextAllowsSafeLinksAndEscapesText(t *testing.T) {
	got, err := SanitizeRichText(`<p><a href="https://example.org/a?x=1&amp;y=2">link</a> &amp; safe</p>`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `href="https://example.org/a?x=1&amp;y=2"`) || !strings.Contains(got, "&amp; safe") {
		t.Fatalf("safe link/text encoding incorrect: %s", got)
	}
}

func TestSanitizeRichTextRejectsOversizedContent(t *testing.T) {
	if _, err := SanitizeRichText(strings.Repeat("x", MaxRichTextBytes+1)); err == nil {
		t.Fatal("expected oversized content to be rejected")
	}
}
