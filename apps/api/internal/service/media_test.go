package service

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	"golang.org/x/image/webp"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/domain"
)

func TestMediaUploadRoleAndPurposePolicy(t *testing.T) {
	tests := []struct {
		role    domain.Role
		purpose MediaPurpose
		allowed bool
	}{
		{domain.RoleSuperAdmin, MediaLandingHero, true},
		{domain.RoleBPH, MediaLandingBranding, true},
		{domain.RoleBPH, MediaAnnouncementImage, true},
		{domain.RoleKepala, MediaArticleCover, true},
		{domain.RoleKepala, MediaGalleryPreview, true},
		{domain.RoleKepala, MediaLandingHero, false},
		{domain.RoleKepala, MediaLandingBranding, false},
		{domain.RoleKepala, MediaMemberPhoto, false},
		{domain.RoleBPH, "UNBOUNDED", false},
		{"UNKNOWN", MediaArticleCover, false},
	}
	for _, test := range tests {
		t.Run(string(test.role)+"/"+string(test.purpose), func(t *testing.T) {
			if got := mediaPurposeAllowed(test.role, test.purpose); got != test.allowed {
				t.Fatalf("mediaPurposeAllowed(%q,%q)=%v want %v", test.role, test.purpose, got, test.allowed)
			}
		})
	}
}

func TestProcessMediaImageSupportedFormats(t *testing.T) {
	fixtures := testImageFixtures(t)
	tests := []struct {
		name       string
		mime       string
		data       []byte
		storedMIME string
		extension  string
	}{
		{"photo.jpg", "image/jpeg", fixtures[".jpg"], "image/jpeg", ".jpg"},
		{"photo.jpeg", "image/jpeg", fixtures[".jpg"], "image/jpeg", ".jpeg"},
		{"photo.jfif", "image/jpeg", fixtures[".jpg"], "image/jpeg", ".jfif"},
		{"photo.png", "image/png", fixtures[".png"], "image/png", ".png"},
		{"photo.webp", "image/webp", fixtures[".webp"], "image/webp", ".webp"},
		{"photo.gif", "image/gif", fixtures[".gif"], "image/gif", ".gif"},
		{"photo.tif", "image/tiff", fixtures[".tif"], "image/png", ".png"},
		{"photo.tiff", "image/tiff", fixtures[".tif"], "image/png", ".png"},
		{"photo.bmp", "image/bmp", fixtures[".bmp"], "image/png", ".png"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			processed, err := processMediaImage(test.data, filepath.Ext(test.name), test.mime, 5<<20)
			if err != nil {
				t.Fatalf("process image: %v", err)
			}
			if processed.storedMIME != test.storedMIME || processed.storedExt != test.extension {
				t.Fatalf("stored type=%s%s want %s%s", processed.storedMIME, processed.storedExt, test.storedMIME, test.extension)
			}
			var config image.Config
			if processed.storedMIME == "image/webp" {
				config, err = webp.DecodeConfig(bytes.NewReader(processed.storedBytes))
			} else {
				config, _, err = image.DecodeConfig(bytes.NewReader(processed.storedBytes))
			}
			if err != nil || config.Width != processed.width || config.Height != processed.height {
				t.Fatalf("stored output did not decode with expected dimensions: config=%+v err=%v", config, err)
			}
		})
	}
}

func TestProcessMediaImageRejectsUnsafeOrInvalidInputs(t *testing.T) {
	fixtures := testImageFixtures(t)
	animated := animatedGIFFixture(t)
	largePNG := encodeTestPNG(t, image.NewRGBA(image.Rect(0, 0, maxMediaDimension+1, 1)))
	tests := []struct {
		name      string
		mime      string
		data      []byte
		limit     int64
		wantError error
	}{
		{"fake jpg", "image/jpeg", []byte("not an image"), 5 << 20, ErrMediaCorrupt},
		{"fake tiff", "image/tiff", []byte("not a tiff"), 5 << 20, ErrMediaCorrupt},
		{"corrupt gif", "image/gif", []byte("GIF89a-broken"), 5 << 20, ErrMediaCorrupt},
		{"oversized bmp", "image/bmp", bytes.Repeat([]byte("x"), (5<<20)+1), 5 << 20, ErrMediaTooLarge},
		{"extreme dimensions", "image/png", largePNG, 5 << 20, ErrMediaDimensions},
		{"svg", "image/svg+xml", []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`), 5 << 20, ErrMediaUnsupported},
		{"unsupported extension", "application/octet-stream", fixtures[".png"], 5 << 20, ErrMediaUnsupported},
		{"mismatched mime", "image/jpeg", fixtures[".png"], 5 << 20, ErrMediaUnsupported},
		{"animated gif", "image/gif", animated, 5 << 20, ErrMediaAnimationUnsupported},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := processMediaImage(test.data, map[string]string{
				"fake jpg": ".jpg", "fake tiff": ".tiff", "corrupt gif": ".gif", "oversized bmp": ".bmp",
				"extreme dimensions": ".png", "svg": ".svg", "unsupported extension": ".svg", "mismatched mime": ".png", "animated gif": ".gif",
			}[test.name], test.mime, test.limit)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("error=%v want %v", err, test.wantError)
			}
		})
	}
}

func testImageFixtures(t *testing.T) map[string][]byte {
	t.Helper()
	imageValue := image.NewRGBA(image.Rect(0, 0, 2, 2))
	imageValue.Set(0, 0, color.RGBA{R: 20, G: 40, B: 60, A: 255})
	fixtures := map[string][]byte{}
	encode := func(extension string, writer func(*bytes.Buffer) error) {
		var buffer bytes.Buffer
		if err := writer(&buffer); err != nil {
			t.Fatalf("encode %s fixture: %v", extension, err)
		}
		fixtures[extension] = buffer.Bytes()
	}
	encode(".jpg", func(buffer *bytes.Buffer) error { return jpeg.Encode(buffer, imageValue, nil) })
	encode(".png", func(buffer *bytes.Buffer) error { return png.Encode(buffer, imageValue) })
	encode(".gif", func(buffer *bytes.Buffer) error { return gif.Encode(buffer, imageValue, nil) })
	encode(".tif", func(buffer *bytes.Buffer) error { return tiff.Encode(buffer, imageValue, nil) })
	encode(".bmp", func(buffer *bytes.Buffer) error { return bmp.Encode(buffer, imageValue) })
	webpData, err := os.ReadFile(filepath.Join("testdata", "webp-sample.webp"))
	if err != nil {
		t.Fatal(err)
	}
	fixtures[".webp"] = webpData
	return fixtures
}

func encodeTestPNG(t *testing.T, value image.Image) []byte {
	t.Helper()
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, value); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func animatedGIFFixture(t *testing.T) []byte {
	t.Helper()
	frame := image.NewPaletted(image.Rect(0, 0, 1, 1), color.Palette{color.Black, color.White})
	frame.SetColorIndex(0, 0, 1)
	var buffer bytes.Buffer
	if err := gif.EncodeAll(&buffer, &gif.GIF{
		Image: []*image.Paletted{frame, frame}, Delay: []int{1, 1},
		Config: image.Config{Width: 1, Height: 1, ColorModel: frame.Palette},
	}); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func TestMediaFilenameAndDimensionValidation(t *testing.T) {
	if got := safeMediaFilename(`..\..\cover.png`); got != "cover.png" {
		t.Fatalf("filename sanitization = %q", got)
	}
	for _, name := range []string{"", "..", strings.Repeat("a", 256)} {
		if got := safeMediaFilename(name); got != "" {
			t.Errorf("unsafe filename %q became %q", name, got)
		}
	}
	if got := safeMediaFilename(`C:\cover.png`); got != "cover.png" {
		t.Errorf("Windows path sanitization = %q", got)
	}
	for _, test := range []struct {
		width, height int
		valid         bool
	}{{1, 1, true}, {8000, 1500, true}, {8000, 1501, false}, {0, 100, false}, {8001, 1, false}} {
		if got := validMediaDimensions(test.width, test.height); got != test.valid {
			t.Errorf("validMediaDimensions(%d,%d)=%v want %v", test.width, test.height, got, test.valid)
		}
	}
}
