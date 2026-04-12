package sws_test

import (
	"path/filepath"
	"strings"
	"testing"

	"sbl/internal/sws"
	"sbl/internal/testutil"
)

func TestGenerateIncludesAliasRedirects(t *testing.T) {
	root := testutil.CopyFixture(t, "site-basic")

	config, err := sws.Generate(root, filepath.Join(root, "public"), map[string]string{
		"/old/": "/posts/hello-world/",
	})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	if !strings.Contains(config, `source = "/old/"`) {
		t.Fatalf("generated config missing alias redirect: %s", config)
	}
	if !strings.Contains(config, `destination = "/posts/hello-world/"`) {
		t.Fatalf("generated config missing alias destination: %s", config)
	}
	if !strings.Contains(config, `source = "/posts/{*}/index.html"`) {
		t.Fatalf("generated config missing structural redirect: %s", config)
	}
}
