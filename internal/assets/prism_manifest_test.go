package assets

import (
	"slices"
	"testing"
)

func TestRequiredPrismComponentLanguagesResolvesAliasesAndDependencies(t *testing.T) {
	available := map[string][]byte{
		prismComponentPath("clike"):      {},
		prismComponentPath("javascript"): {},
		prismComponentPath("jsx"):        {},
		prismComponentPath("markup"):     {},
		prismComponentPath("tsx"):        {},
		prismComponentPath("typescript"): {},
	}

	languages, err := requiredPrismComponentLanguages([]string{"tsx", "js"}, available)
	if err != nil {
		t.Fatalf("requiredPrismComponentLanguages returned error: %v", err)
	}

	want := []string{"clike", "javascript", "markup", "jsx", "typescript", "tsx"}
	if !slices.Equal(languages, want) {
		t.Fatalf("unexpected Prism language closure: got %v want %v", languages, want)
	}
}

func TestRequiredPrismComponentLanguagesRejectsMissingKnownDependency(t *testing.T) {
	available := map[string][]byte{
		prismComponentPath("javascript"): {},
		prismComponentPath("typescript"): {},
	}

	_, err := requiredPrismComponentLanguages([]string{"ts"}, available)
	if err == nil {
		t.Fatal("expected missing known dependency error")
	}
}
