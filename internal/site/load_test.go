package site_test

import (
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
	if len(cfg.Navigation) != 1 || cfg.Navigation[0].URL != "/archive/" {
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
