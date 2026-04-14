package render_test

import (
	"strings"
	"testing"
	"time"

	"sbl/internal/content"
	"sbl/internal/render"
)

func TestRenderPostBodyLeavesCodeBlocksForClientHighlighting(t *testing.T) {
	post := &content.Post{
		Slug:         "code-post",
		SourcePath:   "content/posts/code-post/index.md",
		Title:        "Code Post",
		Date:         time.Date(2026, 4, 14, 0, 0, 0, 0, time.UTC),
		MarkdownBody: "```go\npackage main\n```",
	}

	bodyHTML, generatedFiles, _, features, err := render.RenderPostBody(post, map[string]string{})
	if err != nil {
		t.Fatalf("RenderPostBody returned error: %v", err)
	}
	if len(generatedFiles) != 0 {
		t.Fatalf("expected no generated files, got %d", len(generatedFiles))
	}

	html := string(bodyHTML)
	if !strings.Contains(html, `<code class="language-go">`) {
		t.Fatalf("expected language class in rendered code block: %s", html)
	}
	if strings.Contains(html, `class="chroma"`) {
		t.Fatalf("expected server-side highlighting markup to be absent: %s", html)
	}
	if !features.NeedsCodeHighlight {
		t.Fatalf("expected code highlight feature to be enabled")
	}
	if features.NeedsMath {
		t.Fatalf("expected math feature to be disabled")
	}
	if features.NeedsMermaid {
		t.Fatalf("expected Mermaid feature to be disabled")
	}
}
