package service

import "testing"

func TestValidActionURL(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "fragment", value: "#berita", want: true},
		{name: "internal route", value: "/tentang", want: true},
		{name: "internal route with query", value: "/berita?page=2", want: true},
		{name: "secure external", value: "https://example.test/path", want: true},
		{name: "insecure external", value: "http://example.test", want: false},
		{name: "javascript scheme", value: "javascript:alert(1)", want: false},
		{name: "protocol relative", value: "//example.test/path", want: false},
		{name: "backslash route", value: `/\\example.test`, want: false},
		{name: "empty fragment", value: "#", want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := validActionURL(test.value); got != test.want {
				t.Fatalf("validActionURL(%q) = %t, want %t", test.value, got, test.want)
			}
		})
	}
}

func TestBrandingPresentationValidation(t *testing.T) {
	for _, test := range []struct {
		value string
		valid bool
	}{
		{value: "original", valid: true},
		{value: "rounded", valid: true},
		{value: "circle", valid: true},
		{value: "square", valid: false},
	} {
		if got := validLogoShape(test.value); got != test.valid {
			t.Errorf("validLogoShape(%q) = %t, want %t", test.value, got, test.valid)
		}
	}
	for _, test := range []struct {
		value string
		valid bool
	}{
		{value: "contain", valid: true},
		{value: "cover", valid: true},
		{value: "fill", valid: false},
	} {
		if got := validLogoFit(test.value); got != test.valid {
			t.Errorf("validLogoFit(%q) = %t, want %t", test.value, got, test.valid)
		}
	}
}

func TestValidateSocialLinks(t *testing.T) {
	tests := []struct {
		name  string
		links []SocialLink
		valid bool
	}{
		{name: "HTTPS and future platform", links: []SocialLink{{Platform: "instagram", Label: "Instagram", URL: "https://instagram.com/bem", Active: true}, {Platform: "future-platform", Label: "Kanal", URL: "https://example.test/", Active: false}}, valid: true},
		{name: "HTTP rejected", links: []SocialLink{{Platform: "instagram", Label: "Instagram", URL: "http://instagram.com/bem", Active: true}}},
		{name: "duplicate platform rejected", links: []SocialLink{{Platform: "x", Label: "X", URL: "https://x.com/bem", Active: true}, {Platform: "x", Label: "Twitter", URL: "https://x.com/bem", Active: true}}},
		{name: "active empty URL rejected", links: []SocialLink{{Platform: "website", Label: "Website", Active: true}}},
		{name: "inactive empty URL allowed", links: []SocialLink{{Platform: "website", Label: "Website", Active: false}}, valid: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateSocialLinks(test.links)
			if (err == nil) != test.valid {
				t.Fatalf("validateSocialLinks() error=%v, valid=%t", err, test.valid)
			}
		})
	}
}

func TestLegacySocialLinksMigrationCompatibility(t *testing.T) {
	instagram := "https://instagram.com/bem"
	legacy := legacySocialLinks(LandingSettings{InstagramURL: &instagram})
	if len(legacy) != 1 || legacy[0].Platform != "instagram" || legacy[0].URL != instagram || !legacy[0].Active {
		t.Fatalf("legacy social URL did not map to the generic model: %#v", legacy)
	}
	settings := LandingSettings{SocialLinks: []SocialLink{
		{Platform: "instagram", Label: "Instagram", URL: instagram, Active: true},
		{Platform: "tiktok", Label: "TikTok", URL: "https://tiktok.com/@bem", Active: false},
	}}
	syncLegacySocialURLs(&settings)
	if settings.InstagramURL == nil || *settings.InstagramURL != instagram || settings.TikTokURL != nil {
		t.Fatalf("legacy URL compatibility fields not synchronized: %#v", settings)
	}
}
