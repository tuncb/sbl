package app

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sbl/internal/console"
	"sbl/internal/testutil"
)

func TestScanLiveInputsIgnoresGeneratedDeployConfig(t *testing.T) {
	root := testutil.CopyFixture(t, "site-basic")

	before, err := scanLiveInputs(root)
	if err != nil {
		t.Fatalf("scanLiveInputs returned error: %v", err)
	}

	deployDir := filepath.Join(root, "deploy")
	if err := os.MkdirAll(deployDir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(deployDir, "sws.toml"), []byte("root = \"./public\"\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	after, err := scanLiveInputs(root)
	if err != nil {
		t.Fatalf("scanLiveInputs returned error: %v", err)
	}

	if _, exists := after["deploy/sws.toml"]; exists {
		t.Fatal("generated deploy config should not be part of the watched input snapshot")
	}
	if before["deploy/sws.base.toml"] != after["deploy/sws.base.toml"] {
		t.Fatal("generated deploy config should not affect the watched deploy override input")
	}
}

func TestRunLiveBuildPrintsColorCodedError(t *testing.T) {
	var stderr bytes.Buffer

	runLiveBuild(io.Discard, &stderr, BuildOptions{
		SiteRoot: t.TempDir(),
	}, nil, "")

	output := stderr.String()
	if !strings.HasPrefix(output, console.ErrorPrefix+"live build failed: ") {
		t.Fatalf("stderr = %q, want color-coded live build error", output)
	}
	if !strings.Contains(output, "site config file is required") {
		t.Fatalf("stderr = %q, want build failure detail", output)
	}
}
