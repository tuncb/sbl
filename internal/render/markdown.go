package render

import (
	"bytes"
	"html/template"
	"strings"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	htmlrender "github.com/yuin/goldmark/renderer/html"

	"sbl/internal/assets"
	"sbl/internal/content"
)

var markdownEngine = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		highlighting.NewHighlighting(
			highlighting.WithFormatOptions(
				chromahtml.WithClasses(true),
			),
		),
	),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
	),
	goldmark.WithRendererOptions(
		htmlrender.WithHardWraps(),
	),
)

func RenderPostBody(post *content.Post, localAssetURLs map[string]string) (template.HTML, []assets.File, int, error) {
	return renderDocumentBody("posts", post.Slug, post.MarkdownBody, localAssetURLs)
}

func RenderPageBody(page *content.Page, localAssetURLs map[string]string) (template.HTML, []assets.File, int, error) {
	return renderDocumentBody("pages", page.Slug, page.MarkdownBody, localAssetURLs)
}

func renderDocumentBody(section, slug, markdown string, localAssetURLs map[string]string) (template.HTML, []assets.File, int, error) {
	markdownInput, mermaidBlocks, err := ExtractMermaid(markdown)
	if err != nil {
		return "", nil, 0, err
	}

	markdownInput, displayMathBlocks, err := ExtractDisplayMath(markdownInput)
	if err != nil {
		return "", nil, 0, err
	}

	markdownInput, inlineMathBlocks, err := ExtractInlineMath(markdownInput)
	if err != nil {
		return "", nil, 0, err
	}

	var buffer bytes.Buffer
	if err := markdownEngine.Convert([]byte(markdownInput), &buffer); err != nil {
		return "", nil, 0, err
	}

	htmlFragment := buffer.String()
	htmlFragment, err = ReplaceDisplayMathPlaceholders(htmlFragment, displayMathBlocks)
	if err != nil {
		return "", nil, 0, err
	}

	htmlFragment, err = ReplaceInlineMathPlaceholders(htmlFragment, inlineMathBlocks)
	if err != nil {
		return "", nil, 0, err
	}

	htmlFragment, generatedFiles, err := InjectMermaid(section, slug, htmlFragment, mermaidBlocks)
	if err != nil {
		return "", nil, 0, err
	}

	htmlFragment, err = RewriteAssetLinks(htmlFragment, localAssetURLs)
	if err != nil {
		return "", nil, 0, err
	}

	return template.HTML(htmlFragment), generatedFiles, estimateReadingTime(markdown), nil
}

func estimateReadingTime(markdown string) int {
	words := len(strings.Fields(markdown))
	minutes := (words + 199) / 200
	if minutes < 1 {
		return 1
	}
	return minutes
}
