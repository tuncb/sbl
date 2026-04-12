package assets_test

import (
	"strings"
	"testing"

	"sbl/internal/assets"
)

func TestFingerprintPath(t *testing.T) {
	got := assets.FingerprintPath("posts/hello-world/layout.svg", []byte("fixture"))

	if !strings.HasPrefix(got, "posts/hello-world/layout.") {
		t.Fatalf("unexpected prefix: %q", got)
	}
	if !strings.HasSuffix(got, ".svg") {
		t.Fatalf("unexpected suffix: %q", got)
	}
}
