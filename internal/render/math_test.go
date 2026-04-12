package render_test

import (
	"strings"
	"testing"

	"sbl/internal/render"
)

func TestExtractDisplayMathRejectsUnterminatedBlock(t *testing.T) {
	_, _, err := render.ExtractDisplayMath("$$\nx + 1")
	if err == nil || !strings.Contains(err.Error(), "unterminated display math block") {
		t.Fatalf("expected unterminated display math error, got: %v", err)
	}
}

func TestRenderInlineMathRejectsUnterminatedExpression(t *testing.T) {
	_, err := render.RenderInlineMath("<p>Inline \\(x + 1</p>")
	if err == nil || !strings.Contains(err.Error(), "unterminated inline math delimiter") {
		t.Fatalf("expected inline math error, got: %v", err)
	}
}
