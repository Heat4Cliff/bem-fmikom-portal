package router

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	"golang.org/x/image/webp"
)

type contentMediaFixtures struct {
	AnnouncementImage          string
	GalleryPreview             string
	LandingHero                string
	LandingBranding            string
	LandingBrandingReplacement string
	ArticleCovers              []articleCoverFixture
}

type articleCoverFixture struct {
	Filename        string
	OriginalMIME    string
	StoredMIME      string
	PublicURL       string
	NormalizedBytes []byte
}

func runMediaIntegration(t *testing.T, ctx context.Context, db *sql.DB, engine http.Handler, superToken, bphToken, kepalaToken, root string) contentMediaFixtures {
	t.Helper()
	var pngData bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 20, G: 40, B: 60, A: 255})
	if err := png.Encode(&pngData, img); err != nil {
		t.Fatal(err)
	}
	uploadWithDetails := func(token, purpose, filename, contentType string, file []byte) *httptest.ResponseRecorder {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		_ = writer.WriteField("purpose", purpose)
		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", mime.FormatMediaType("form-data", map[string]string{"name": "file", "filename": filename}))
		header.Set("Content-Type", contentType)
		part, err := writer.CreatePart(header)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write(file); err != nil {
			t.Fatal(err)
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/media", &body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		response := httptest.NewRecorder()
		engine.ServeHTTP(response, req)
		return response
	}
	upload := func(token, purpose string, file []byte) *httptest.ResponseRecorder {
		return uploadWithDetails(token, purpose, "portrait.png", "image/png", file)
	}
	if response := upload("", "ARTICLE_COVER", pngData.Bytes()); response.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous media upload=%d %s", response.Code, response.Body.String())
	}
	heroUpload := upload(superToken, "LANDING_HERO", pngData.Bytes())
	if heroUpload.Code != http.StatusCreated {
		t.Fatalf("SUPER_ADMIN media upload=%d %s", heroUpload.Code, heroUpload.Body.String())
	}
	brandingUpload := upload(superToken, "LANDING_BRANDING", pngData.Bytes())
	if brandingUpload.Code != http.StatusCreated {
		t.Fatalf("SUPER_ADMIN branding media upload=%d %s", brandingUpload.Code, brandingUpload.Body.String())
	}
	replacementPixels := image.NewRGBA(image.Rect(0, 0, 2, 2))
	replacementPixels.Set(0, 0, color.RGBA{R: 80, G: 120, B: 160, A: 255})
	var replacementPNG bytes.Buffer
	if err := png.Encode(&replacementPNG, replacementPixels); err != nil {
		t.Fatal(err)
	}
	brandingReplacementUpload := upload(superToken, "LANDING_BRANDING", replacementPNG.Bytes())
	if brandingReplacementUpload.Code != http.StatusCreated {
		t.Fatalf("SUPER_ADMIN replacement branding media upload=%d %s", brandingReplacementUpload.Code, brandingReplacementUpload.Body.String())
	}
	bphUpload := upload(bphToken, "ANNOUNCEMENT_IMAGE", pngData.Bytes())
	if bphUpload.Code != http.StatusCreated {
		t.Fatalf("BPH media upload=%d %s", bphUpload.Code, bphUpload.Body.String())
	}
	galleryUpload := upload(bphToken, "GALLERY_PREVIEW", pngData.Bytes())
	if galleryUpload.Code != http.StatusCreated {
		t.Fatalf("BPH gallery preview media upload=%d %s", galleryUpload.Code, galleryUpload.Body.String())
	}
	kepalaUpload := upload(kepalaToken, "ARTICLE_COVER", pngData.Bytes())
	if kepalaUpload.Code != http.StatusCreated {
		t.Fatalf("KEPALA_BIDANG scoped media upload=%d %s", kepalaUpload.Code, kepalaUpload.Body.String())
	}
	if response := upload(kepalaToken, "LANDING_HERO", pngData.Bytes()); response.Code != http.StatusForbidden {
		t.Fatalf("KEPALA_BIDANG was allowed system media purpose: %d %s", response.Code, response.Body.String())
	}
	if response := upload(superToken, "ARTICLE_COVER", []byte("<html>not an image</html>")); response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("HTML renamed PNG was accepted: %d %s", response.Code, response.Body.String())
	}
	var payload struct {
		Data struct {
			ID        int64  `json:"id"`
			PublicURL string `json:"public_url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(kepalaUpload.Body.Bytes(), &payload); err != nil || payload.Data.ID < 1 {
		t.Fatalf("parse media upload response: %v %s", err, kepalaUpload.Body.String())
	}
	read := httptest.NewRecorder()
	engine.ServeHTTP(read, httptest.NewRequest(http.MethodGet, payload.Data.PublicURL, nil))
	if read.Code != http.StatusOK || read.Header().Get("Content-Type") != "image/png" || read.Header().Get("X-Content-Type-Options") != "nosniff" || read.Header().Get("Cache-Control") != "public, max-age=31536000, immutable" || !bytes.Equal(read.Body.Bytes(), pngData.Bytes()) {
		t.Fatalf("public media read failed: status=%d headers=%v size=%d", read.Code, read.Header(), read.Body.Len())
	}
	var storageKey, digest string
	if err := db.QueryRowContext(ctx, "SELECT storage_key,sha256 FROM media_assets WHERE id=?", payload.Data.ID).Scan(&storageKey, &digest); err != nil {
		t.Fatal(err)
	}
	if filepath.IsAbs(storageKey) || filepath.Base(storageKey) != storageKey || len(digest) != 64 {
		t.Fatalf("media metadata contains unsafe key/hash: %q %q", storageKey, digest)
	}
	if _, err := os.Stat(filepath.Join(root, storageKey)); err != nil {
		t.Fatalf("public object was not persisted under opaque key: %v", err)
	}
	articleCovers := verifyImageFormats(t, ctx, db, engine, superToken, uploadWithDetails)
	verifyInvalidImageUploads(t, superToken, uploadWithDetails)
	var announcement, gallery struct {
		Data struct {
			PublicURL string `json:"public_url"`
		} `json:"data"`
	}
	var hero, branding struct {
		Data struct {
			PublicURL string `json:"public_url"`
		} `json:"data"`
	}
	var brandingReplacement struct {
		Data struct {
			PublicURL string `json:"public_url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(heroUpload.Body.Bytes(), &hero); err != nil || hero.Data.PublicURL == "" {
		t.Fatalf("parse landing hero media response: %v %s", err, heroUpload.Body.String())
	}
	if err := json.Unmarshal(brandingUpload.Body.Bytes(), &branding); err != nil || branding.Data.PublicURL == "" {
		t.Fatalf("parse landing branding media response: %v %s", err, brandingUpload.Body.String())
	}
	if err := json.Unmarshal(brandingReplacementUpload.Body.Bytes(), &brandingReplacement); err != nil || brandingReplacement.Data.PublicURL == "" || brandingReplacement.Data.PublicURL == branding.Data.PublicURL {
		t.Fatalf("parse replacement landing branding media response: %v %s", err, brandingReplacementUpload.Body.String())
	}
	if err := json.Unmarshal(bphUpload.Body.Bytes(), &announcement); err != nil || announcement.Data.PublicURL == "" {
		t.Fatalf("parse announcement media response: %v %s", err, bphUpload.Body.String())
	}
	if err := json.Unmarshal(galleryUpload.Body.Bytes(), &gallery); err != nil || gallery.Data.PublicURL == "" {
		t.Fatalf("parse gallery media response: %v %s", err, galleryUpload.Body.String())
	}
	return contentMediaFixtures{AnnouncementImage: announcement.Data.PublicURL, GalleryPreview: gallery.Data.PublicURL, LandingHero: hero.Data.PublicURL, LandingBranding: branding.Data.PublicURL, LandingBrandingReplacement: brandingReplacement.Data.PublicURL, ArticleCovers: articleCovers}
}

type mediaUploadFunc func(token, purpose, filename, contentType string, file []byte) *httptest.ResponseRecorder

type imageFormatFixture struct {
	filename   string
	inputMIME  string
	storedMIME string
	data       []byte
}

func verifyImageFormats(t *testing.T, ctx context.Context, db *sql.DB, engine http.Handler, token string, upload mediaUploadFunc) []articleCoverFixture {
	t.Helper()
	fixtures := imageFormatFixtures(t)
	articleCovers := make([]articleCoverFixture, 0, len(fixtures))
	for _, fixture := range fixtures {
		response := upload(token, "ARTICLE_COVER", fixture.filename, fixture.inputMIME, fixture.data)
		if response.Code != http.StatusCreated {
			t.Fatalf("upload %s returned %d: %s", fixture.filename, response.Code, response.Body.String())
		}
		var payload struct {
			Data struct {
				ID               int64  `json:"id"`
				OriginalFilename string `json:"original_filename"`
				OriginalMIME     string `json:"original_mime_type"`
				MIME             string `json:"mime_type"`
				StoredMIME       string `json:"stored_mime_type"`
				FileExtension    string `json:"file_extension"`
				PublicURL        string `json:"public_url"`
			} `json:"data"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil || payload.Data.ID < 1 {
			t.Fatalf("decode %s media response: %v %s", fixture.filename, err, response.Body.String())
		}
		wantExtension := filepath.Ext(fixture.filename)
		if fixture.inputMIME == "image/tiff" || fixture.inputMIME == "image/bmp" {
			wantExtension = ".png"
		}
		if payload.Data.OriginalMIME != fixture.inputMIME || payload.Data.StoredMIME != fixture.storedMIME || payload.Data.MIME != fixture.storedMIME || payload.Data.FileExtension != wantExtension {
			t.Fatalf("%s metadata original=%s stored=%s mime=%s ext=%s", fixture.filename, payload.Data.OriginalMIME, payload.Data.StoredMIME, payload.Data.MIME, payload.Data.FileExtension)
		}
		read := httptest.NewRecorder()
		engine.ServeHTTP(read, httptest.NewRequest(http.MethodGet, payload.Data.PublicURL, nil))
		if read.Code != http.StatusOK || read.Header().Get("Content-Type") != fixture.storedMIME {
			t.Fatalf("read normalized %s output: status=%d content-type=%q", fixture.filename, read.Code, read.Header().Get("Content-Type"))
		}
		if fixture.storedMIME == "image/webp" {
			if _, err := webp.DecodeConfig(bytes.NewReader(read.Body.Bytes())); err != nil {
				t.Fatalf("stored WEBP output is invalid: %v", err)
			}
		} else if _, _, err := image.DecodeConfig(bytes.NewReader(read.Body.Bytes())); err != nil {
			t.Fatalf("stored %s output is invalid: %v", fixture.filename, err)
		}
		var storedOriginalMIME, storedMIME, storedExtension, storageKey, originalFilename string
		if err := db.QueryRowContext(ctx, `SELECT original_mime_type,mime_type,file_extension,storage_key,original_filename FROM media_assets WHERE id=?`, payload.Data.ID).Scan(&storedOriginalMIME, &storedMIME, &storedExtension, &storageKey, &originalFilename); err != nil {
			t.Fatalf("load %s media metadata: %v", fixture.filename, err)
		}
		if storedOriginalMIME != fixture.inputMIME || storedMIME != fixture.storedMIME || storedExtension != wantExtension || originalFilename != fixture.filename || filepath.Base(storageKey) != storageKey {
			t.Fatalf("persisted %s metadata original=%s stored=%s ext=%s filename=%s key=%q", fixture.filename, storedOriginalMIME, storedMIME, storedExtension, originalFilename, storageKey)
		}
		articleCovers = append(articleCovers, articleCoverFixture{
			Filename: fixture.filename, OriginalMIME: fixture.inputMIME, StoredMIME: fixture.storedMIME,
			PublicURL: payload.Data.PublicURL, NormalizedBytes: append([]byte(nil), read.Body.Bytes()...),
		})
	}
	return articleCovers
}

func imageFormatFixtures(t *testing.T) []imageFormatFixture {
	t.Helper()
	pixels := image.NewRGBA(image.Rect(0, 0, 2, 2))
	pixels.Set(0, 0, color.RGBA{R: 50, G: 100, B: 150, A: 255})
	encode := func(name string, write func(*bytes.Buffer) error) []byte {
		var buffer bytes.Buffer
		if err := write(&buffer); err != nil {
			t.Fatalf("encode %s fixture: %v", name, err)
		}
		return buffer.Bytes()
	}
	webpBytes, err := os.ReadFile(filepath.Join("..", "..", "service", "testdata", "webp-sample.webp"))
	if err != nil {
		t.Fatal(err)
	}
	return []imageFormatFixture{
		{filename: "article.jpg", inputMIME: "image/jpeg", storedMIME: "image/jpeg", data: encode("jpeg", func(out *bytes.Buffer) error { return jpeg.Encode(out, pixels, nil) })},
		{filename: "article.jfif", inputMIME: "image/jpeg", storedMIME: "image/jpeg", data: encode("jfif", func(out *bytes.Buffer) error { return jpeg.Encode(out, pixels, nil) })},
		{filename: "article.png", inputMIME: "image/png", storedMIME: "image/png", data: encode("png", func(out *bytes.Buffer) error { return png.Encode(out, pixels) })},
		{filename: "article.webp", inputMIME: "image/webp", storedMIME: "image/webp", data: webpBytes},
		{filename: "article.gif", inputMIME: "image/gif", storedMIME: "image/gif", data: encode("gif", func(out *bytes.Buffer) error { return gif.Encode(out, pixels, nil) })},
		{filename: "article.tiff", inputMIME: "image/tiff", storedMIME: "image/png", data: encode("tiff", func(out *bytes.Buffer) error { return tiff.Encode(out, pixels, nil) })},
		{filename: "article.bmp", inputMIME: "image/bmp", storedMIME: "image/png", data: encode("bmp", func(out *bytes.Buffer) error { return bmp.Encode(out, pixels) })},
	}
}

func verifyInvalidImageUploads(t *testing.T, token string, upload mediaUploadFunc) {
	t.Helper()
	largePixels := image.NewRGBA(image.Rect(0, 0, 8001, 1))
	var oversized bytes.Buffer
	oversized.Write(bytes.Repeat([]byte("x"), (5<<20)+1))
	animatedFrame := image.NewPaletted(image.Rect(0, 0, 1, 1), color.Palette{color.Black, color.White})
	var animatedGIF bytes.Buffer
	if err := gif.EncodeAll(&animatedGIF, &gif.GIF{
		Image: []*image.Paletted{animatedFrame, animatedFrame}, Delay: []int{1, 1},
		Config: image.Config{Width: 1, Height: 1, ColorModel: animatedFrame.Palette},
	}); err != nil {
		t.Fatal(err)
	}
	var largePNG bytes.Buffer
	if err := png.Encode(&largePNG, largePixels); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		filename, mime, message string
		data                    []byte
		status                  int
	}{
		{"fake.jpg", "image/jpeg", "File gambar rusak atau tidak dapat dibaca.", []byte("not an image"), http.StatusUnprocessableEntity},
		{"fake.tiff", "image/tiff", "File gambar rusak atau tidak dapat dibaca.", []byte("not a tiff"), http.StatusUnprocessableEntity},
		{"bad.gif", "image/gif", "File gambar rusak atau tidak dapat dibaca.", []byte("GIF89a-broken"), http.StatusUnprocessableEntity},
		{"large.bmp", "image/bmp", "Ukuran gambar melebihi batas 5 MiB.", oversized.Bytes(), http.StatusRequestEntityTooLarge},
		{"huge.png", "image/png", "Dimensi gambar melebihi batas aman.", largePNG.Bytes(), http.StatusUnprocessableEntity},
		{"vector.svg", "image/svg+xml", "Format gambar belum didukung.", []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`), http.StatusUnprocessableEntity},
		{"motion.gif", "image/gif", "GIF animasi belum didukung.", animatedGIF.Bytes(), http.StatusUnprocessableEntity},
	}
	for _, test := range tests {
		t.Run(test.filename, func(t *testing.T) {
			response := upload(token, "ARTICLE_COVER", test.filename, test.mime, test.data)
			if response.Code != test.status || !bytes.Contains(response.Body.Bytes(), []byte(test.message)) {
				t.Fatalf("invalid upload response=%d %s, want status %d and message %q", response.Code, response.Body.String(), test.status, test.message)
			}
		})
	}
	pathTraversal := upload(token, "ARTICLE_COVER", `..\..\safe.jpg`, "image/jpeg", imageFormatFixtures(t)[0].data)
	if pathTraversal.Code != http.StatusCreated {
		t.Fatalf("path traversal filename did not safely upload after sanitization: %d %s", pathTraversal.Code, pathTraversal.Body.String())
	}
	var sanitized struct {
		Data struct {
			OriginalFilename string `json:"original_filename"`
		} `json:"data"`
	}
	if err := json.Unmarshal(pathTraversal.Body.Bytes(), &sanitized); err != nil || sanitized.Data.OriginalFilename != "safe.jpg" {
		t.Fatalf("path traversal filename was not reduced to a basename: err=%v filename=%q", err, sanitized.Data.OriginalFilename)
	}
}
