package main

import (
	"bytes"
	"testing"

	"sbl/internal/app"
)

func TestRunBuildAcceptsSiteRootBeforeFlags(t *testing.T) {
	original := buildFn
	t.Cleanup(func() {
		buildFn = original
	})

	var got app.BuildOptions
	buildFn = func(opts app.BuildOptions) error {
		got = opts
		return nil
	}

	code := run([]string{
		"build",
		"./site",
		"--clean",
		"--out", "dist",
		"--base-url", "https://example.com",
		"--include-drafts",
	})
	if code != 0 {
		t.Fatalf("run returned %d", code)
	}
	if got.SiteRoot != "./site" {
		t.Fatalf("SiteRoot = %q, want %q", got.SiteRoot, "./site")
	}
	if got.OutputDir != "dist" {
		t.Fatalf("OutputDir = %q, want %q", got.OutputDir, "dist")
	}
	if got.BaseURL != "https://example.com" {
		t.Fatalf("BaseURL = %q, want %q", got.BaseURL, "https://example.com")
	}
	if !got.IncludeDrafts {
		t.Fatal("IncludeDrafts = false, want true")
	}
	if !got.Clean {
		t.Fatal("Clean = false, want true")
	}
}

func TestRunValidateAcceptsSiteRootBeforeFlags(t *testing.T) {
	original := validateFn
	t.Cleanup(func() {
		validateFn = original
	})

	var got app.ValidateOptions
	validateFn = func(opts app.ValidateOptions) error {
		got = opts
		return nil
	}

	code := run([]string{
		"validate",
		"./site",
		"--base-url", "https://example.com",
		"--include-drafts",
	})
	if code != 0 {
		t.Fatalf("run returned %d", code)
	}
	if got.SiteRoot != "./site" {
		t.Fatalf("SiteRoot = %q, want %q", got.SiteRoot, "./site")
	}
	if got.BaseURL != "https://example.com" {
		t.Fatalf("BaseURL = %q, want %q", got.BaseURL, "https://example.com")
	}
	if !got.IncludeDrafts {
		t.Fatal("IncludeDrafts = false, want true")
	}
}

func TestRunVersionPrintsVersion(t *testing.T) {
	original := stdout
	t.Cleanup(func() {
		stdout = original
	})

	var got bytes.Buffer
	stdout = &got

	code := run([]string{"version"})
	if code != 0 {
		t.Fatalf("run returned %d", code)
	}

	if got.String() != version+"\n" {
		t.Fatalf("stdout = %q, want %q", got.String(), version+"\n")
	}
}
