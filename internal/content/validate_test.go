package content_test

import (
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

	if _, err := content.Validate(posts, false); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
}

func TestValidateReportsMissingAsset(t *testing.T) {
	root := testutil.CopyFixture(t, "site-invalid-missing-asset")
	posts, err := content.LoadPosts(root)
	if err != nil {
		t.Fatalf("LoadPosts returned error: %v", err)
	}

	_, err = content.Validate(posts, false)
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

	_, err = content.Validate(posts, false)
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

	_, err = content.Validate(posts, false)
	if err == nil || !strings.Contains(err.Error(), `alias collision: /shared/`) {
		t.Fatalf("expected alias collision error, got: %v", err)
	}
}
