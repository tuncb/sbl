package render

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"html"
	"path"
	"regexp"
	"strings"

	"sbl/internal/assets"
)

type MermaidBlock struct {
	Index       int
	Placeholder string
	Source      string
}

func ExtractMermaid(markdown string) (string, []MermaidBlock, error) {
	lines := strings.Split(markdown, "\n")
	var output []string
	blocks := make([]MermaidBlock, 0)
	var fenceChar byte
	var fenceLen int
	inFence := false
	inMermaid := false
	var current []string

	for _, rawLine := range lines {
		line := strings.TrimRight(rawLine, "\r")
		if !inFence {
			matched, char, count, info := parseFence(line)
			if matched {
				inFence = true
				fenceChar = char
				fenceLen = count
				fields := strings.Fields(info)
				inMermaid = len(fields) > 0 && strings.EqualFold(fields[0], "mermaid")
				if !inMermaid {
					output = append(output, line)
				}
				current = current[:0]
				continue
			}
			output = append(output, line)
			continue
		}

		if isFenceClose(line, fenceChar, fenceLen) {
			if inMermaid {
				index := len(blocks) + 1
				placeholder := fmt.Sprintf("SBL_MERMAID_BLOCK_%d", index)
				blocks = append(blocks, MermaidBlock{
					Index:       index,
					Placeholder: placeholder,
					Source:      strings.TrimSpace(strings.Join(current, "\n")),
				})
				output = append(output, placeholder)
			} else {
				output = append(output, line)
			}
			inFence = false
			inMermaid = false
			fenceChar = 0
			fenceLen = 0
			current = current[:0]
			continue
		}

		if inMermaid {
			current = append(current, line)
			continue
		}
		output = append(output, line)
	}

	if inMermaid {
		return "", nil, fmt.Errorf("unterminated mermaid fence")
	}

	return strings.Join(output, "\n"), blocks, nil
}

func InjectMermaid(slug, htmlFragment string, blocks []MermaidBlock) (string, []assets.File, error) {
	generated := make([]assets.File, 0, len(blocks))
	for _, block := range blocks {
		svg, err := RenderMermaidSVG(block.Source)
		if err != nil {
			return "", nil, fmt.Errorf("render Mermaid block %d: %w", block.Index, err)
		}
		file := assets.NewHashedFile(path.Join("posts", slug, fmt.Sprintf("diagram-%d.svg", block.Index)), svg)
		generated = append(generated, file)
		replacement := fmt.Sprintf(`<figure class="diagram"><img src="%s" alt="Diagram %d"></figure>`, html.EscapeString(file.URL), block.Index)
		var replaced bool
		htmlFragment, replaced = replaceParagraphPlaceholder(htmlFragment, block.Placeholder, replacement)
		if !replaced {
			return "", nil, fmt.Errorf("missing Mermaid placeholder %q in rendered HTML", block.Placeholder)
		}
	}
	return htmlFragment, generated, nil
}

func RenderMermaidSVG(source string) ([]byte, error) {
	if strings.TrimSpace(source) == "" {
		return nil, fmt.Errorf("Mermaid diagram is empty")
	}
	lines := strings.Split(strings.TrimSpace(source), "\n")
	height := 72 + (len(lines) * 22)

	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 800 %d" role="img" aria-label="Mermaid diagram">`, height))
	buffer.WriteString(`<rect width="800" height="100%" fill="#fffdf8" stroke="#ddd0c2"/>`)
	buffer.WriteString(`<text x="32" y="48" font-family="monospace" font-size="18" fill="#1e1a17">`)
	for index, line := range lines {
		buffer.WriteString(fmt.Sprintf(`<tspan x="32" y="%d">`, 48+(index*22)))
		if err := xml.EscapeText(&buffer, []byte(line)); err != nil {
			return nil, err
		}
		buffer.WriteString(`</tspan>`)
	}
	buffer.WriteString(`</text></svg>`)
	return buffer.Bytes(), nil
}

func replaceParagraphPlaceholder(htmlFragment, placeholder, replacement string) (string, bool) {
	pattern := regexp.MustCompile(`<p>\s*` + regexp.QuoteMeta(placeholder) + `\s*</p>`)
	if !pattern.MatchString(htmlFragment) {
		return htmlFragment, false
	}
	return pattern.ReplaceAllString(htmlFragment, replacement), true
}
