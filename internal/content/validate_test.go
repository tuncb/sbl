package content_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sbl/internal/content"
	"sbl/internal/testutil"
)

func TestValidateBasicFixturePasses(t *testing.T) {
	root := testutil.CopyFixture(t, "site-basic")
	posts, err := content.LoadPosts(root)
	if err != nil {
		t.Fatalf("LoadPosts returned error: %v", err)
	}
	pages, err := content.LoadPages(root)
	if err != nil {
		t.Fatalf("LoadPages returned error: %v", err)
	}

	if _, err := content.Validate(posts, pages, false); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
}

func TestValidateLayoutRequiresContentDirectory(t *testing.T) {
	root := t.TempDir()

	err := content.ValidateLayout(root)
	if err == nil || !strings.Contains(err.Error(), "site content directory does not exist") {
		t.Fatalf("expected missing content directory error, got: %v", err)
	}
}

func TestLoadPostsRejectsUnexpectedFiles(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "content", "posts"), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "content", "posts", "wrong-layout.md"), []byte("body"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	_, err := content.LoadPosts(root)
	if err == nil || !strings.Contains(err.Error(), "unexpected file in posts directory") {
		t.Fatalf("expected misplaced file error, got: %v", err)
	}
}

func TestLoadPostsRejectsDirectoriesWithoutIndex(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "content", "posts", "wrong-layout"), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	_, err := content.LoadPosts(root)
	if err == nil || !strings.Contains(err.Error(), "missing index.md in posts directory") {
		t.Fatalf("expected missing index error, got: %v", err)
	}
}

func TestValidateReportsMissingAsset(t *testing.T) {
	root := testutil.CopyFixture(t, "site-invalid-missing-asset")
	posts, err := content.LoadPosts(root)
	if err != nil {
		t.Fatalf("LoadPosts returned error: %v", err)
	}
	pages, err := content.LoadPages(root)
	if err != nil {
		t.Fatalf("LoadPages returned error: %v", err)
	}

	_, err = content.Validate(posts, pages, false)
	if err == nil || !strings.Contains(err.Error(), `missing asset "assets/missing.svg"`) {
		t.Fatalf("expected missing asset error, got: %v", err)
	}
}

func TestValidateReportsBrokenLink(t *testing.T) {
	root := testutil.CopyFixture(t, "site-invalid-broken-link")
	posts, err := content.LoadPosts(root)
	if err != nil {
		t.Fatalf("LoadPosts returned error: %v", err)
	}
	pages, err := content.LoadPages(root)
	if err != nil {
		t.Fatalf("LoadPages returned error: %v", err)
	}

	_, err = content.Validate(posts, pages, false)
	if err == nil || !strings.Contains(err.Error(), `broken internal link "/posts/does-not-exist/"`) {
		t.Fatalf("expected broken link error, got: %v", err)
	}
}

func TestValidateReportsAliasCollision(t *testing.T) {
	root := testutil.CopyFixture(t, "site-invalid-alias-collision")
	posts, err := content.LoadPosts(root)
	if err != nil {
		t.Fatalf("LoadPosts returned error: %v", err)
	}
	pages, err := content.LoadPages(root)
	if err != nil {
		t.Fatalf("LoadPages returned error: %v", err)
	}

	_, err = content.Validate(posts, pages, false)
	if err == nil || !strings.Contains(err.Error(), `alias collision: /shared/`) {
		t.Fatalf("expected alias collision error, got: %v", err)
	}
}
