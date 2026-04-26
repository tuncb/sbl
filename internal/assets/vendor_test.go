package assets_test

import (
	"path/filepath"
	"strings"
	"testing"

	"sbl/internal/assets"
)

func TestNormalizePrismThemeDefaultsAndValidates(t *testing.T) {
	theme, err := assets.NormalizePrismTheme("")
	if err != nil {
		t.Fatalf("NormalizePrismTheme returned error: %v", err)
	}
	if theme != assets.DefaultPrismTheme() {
		t.Fatalf("unexpected default Prism theme: %q", theme)
	}

	theme, err = assets.NormalizePrismTheme(" Okaidia ")
	if err != nil {
		t.Fatalf("NormalizePrismTheme returned error: %v", err)
	}
	if theme != "okaidia" {
		t.Fatalf("unexpected normalized Prism theme: %q", theme)
	}
}

func TestNormalizePrismThemeRejectsUnknownTheme(t *testing.T) {
	_, err := assets.NormalizePrismTheme("missing")
	if err == nil {
		t.Fatal("expected unknown Prism theme error")
	}
	if !strings.Contains(err.Error(), `unknown prism_theme "missing"`) {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(err.Error(), "okaidia") || !strings.Contains(err.Error(), "twilight") {
		t.Fatalf("error missing supported theme list: %v", err)
	}
}

func TestBuildVendorFilesPublishesOnlySelectedPrismTheme(t *testing.T) {
	files, vendorAssets, err := assets.BuildVendorFiles(assets.VendorRequest{
		PrismLanguages: []string{"go"},
		PrismTheme:     "okaidia",
	})
	if err != nil {
		t.Fatalf("BuildVendorFiles returned error: %v", err)
	}

	if !strings.HasSuffix(vendorAssets.PrismCSSURL, "/themes/prism-okaidia.min.css") {
		t.Fatalf("unexpected Prism stylesheet URL: %q", vendorAssets.PrismCSSURL)
	}

	themeFiles := map[string]struct{}{}
	for _, file := range files {
		if filepath.ToSlash(filepath.Dir(file.RelPath)) == "assets/vendor/prism-1.30.0/themes" {
			themeFiles[filepath.Base(file.RelPath)] = struct{}{}
		}
	}

	if _, ok := themeFiles["prism-okaidia.min.css"]; !ok {
		t.Fatalf("selected Prism theme was not published: %v", themeFiles)
	}
	if len(themeFiles) != 1 {
		t.Fatalf("expected only selected Prism theme, got: %v", themeFiles)
	}
}
