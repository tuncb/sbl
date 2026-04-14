package main

import (
	"bytes"
	"strings"
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
		"--timings",
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
	if !got.Timings {
		t.Fatal("Timings = false, want true")
	}
}

func TestRunLiveAcceptsSiteRootBeforeFlags(t *testing.T) {
	original := liveFn
	t.Cleanup(func() {
		liveFn = original
	})

	var got app.LiveOptions
	liveFn = func(opts app.LiveOptions) error {
		got = opts
		return nil
	}

	code := run([]string{
		"live",
		"./site",
		"--out", "dist",
		"--base-url", "https://example.com",
		"--include-drafts",
		"--timings",
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
	if !got.Timings {
		t.Fatal("Timings = false, want true")
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
		"--timings",
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
	if !got.Timings {
		t.Fatal("Timings = false, want true")
	}
}

func TestRunVersionFlagPrintsVersion(t *testing.T) {
	original := stdout
	t.Cleanup(func() {
		stdout = original
	})

	var got bytes.Buffer
	stdout = &got

	code := run([]string{"--version"})
	if code != 0 {
		t.Fatalf("run returned %d", code)
	}

	if got.String() != version+"\n" {
		t.Fatalf("stdout = %q, want %q", got.String(), version+"\n")
	}
}

func TestRunShortVersionFlagPrintsVersion(t *testing.T) {
	original := stdout
	t.Cleanup(func() {
		stdout = original
	})

	var got bytes.Buffer
	stdout = &got

	code := run([]string{"-v"})
	if code != 0 {
		t.Fatalf("run returned %d", code)
	}

	if got.String() != version+"\n" {
		t.Fatalf("stdout = %q, want %q", got.String(), version+"\n")
	}
}

func TestRunVersionWithTimingsPrintsTimingSummary(t *testing.T) {
	original := stdout
	t.Cleanup(func() {
		stdout = original
	})

	var got bytes.Buffer
	stdout = &got

	code := run([]string{"version", "--timings"})
	if code != 0 {
		t.Fatalf("run returned %d", code)
	}

	output := got.String()
	if !strings.Contains(output, version+"\n") {
		t.Fatalf("stdout = %q, want version line", output)
	}
	if !strings.Contains(output, "timings:\n  total: ") {
		t.Fatalf("stdout = %q, want timing summary", output)
	}
}
