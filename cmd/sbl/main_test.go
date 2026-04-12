package main

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"

	"sbl/internal/app"
)

func TestRunSetupCommand(t *testing.T) {
	original := setupFn
	t.Cleanup(func() {
		setupFn = original
	})

	var called bool
	setupFn = func(opts app.SetupOptions) error {
		called = true
		if opts.SkipNPM || opts.SkipBrowser {
			t.Fatalf("unexpected flags: %+v", opts)
		}
		return nil
	}

	exitCode := run([]string{"setup"})
	if exitCode != 0 {
		t.Fatalf("unexpected exit code: %d", exitCode)
	}
	if !called {
		t.Fatal("expected setup to be called")
	}
}

func TestRunSetupCommandError(t *testing.T) {
	original := setupFn
	originalStderr := os.Stderr
	t.Cleanup(func() {
		setupFn = original
		os.Stderr = originalStderr
	})

	var stderr bytes.Buffer
	file, err := os.CreateTemp(t.TempDir(), "stderr")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	os.Stderr = file

	setupFn = func(opts app.SetupOptions) error {
		return errors.New("boom")
	}

	exitCode := run([]string{"setup"})
	if exitCode != 1 {
		t.Fatalf("unexpected exit code: %d", exitCode)
	}
	if _, err := file.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := stderr.ReadFrom(file); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stderr.String(), "boom") {
		t.Fatalf("expected error in stderr, got %q", stderr.String())
	}
}
