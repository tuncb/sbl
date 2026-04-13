package site_test

import (
	"path/filepath"
	"strings"
	"testing"

	"sbl/internal/site"
	"sbl/internal/testutil"
)

func TestLoadConfigFromFixture(t *testing.T) {
	root := testutil.CopyFixture(t, "site-basic")

	cfg, err := site.Load(root, "", true)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Title != "Fixture Blog" {
		t.Fatalf("unexpected title: %q", cfg.Title)
	}
	if cfg.BaseURL != "https://example.test" {
		t.Fatalf("unexpected base URL: %q", cfg.BaseURL)
	}
	if len(cfg.Navigation) != 2 || cfg.Navigation[0].URL != "/archive/" || cfg.Navigation[1].URL != "/pages/about/" {
		t.Fatalf("unexpected navigation: %+v", cfg.Navigation)
	}
}

func TestLoadConfigOverrideBaseURL(t *testing.T) {
	root := testutil.CopyFixture(t, "site-basic")

	cfg, err := site.Load(root, "https://override.example.test", true)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.BaseURL != "https://override.example.test" {
		t.Fatalf("unexpected override base URL: %q", cfg.BaseURL)
	}
}

func TestLoadRejectsMissingSiteRoot(t *testing.T) {
	root := filepath.Join(t.TempDir(), "missing-site")

	_, err := site.Load(root, "", false)
	if err == nil || !strings.Contains(err.Error(), "site root does not exist") {
		t.Fatalf("expected missing site root error, got: %v", err)
	}
}

func TestLoadRejectsMissingConfig(t *testing.T) {
	root := t.TempDir()

	_, err := site.Load(root, "", false)
	if err == nil || !strings.Contains(err.Error(), "site config file is required") {
		t.Fatalf("expected missing config error, got: %v", err)
	}
}
