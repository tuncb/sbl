package render

import (
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

func InjectMermaid(section, slug, htmlFragment string, blocks []MermaidBlock) (string, []assets.File, error) {
	svgs, err := renderMermaidDiagrams(blocks)
	if err != nil {
		return "", nil, err
	}
	generated := make([]assets.File, 0, len(blocks))
	for _, block := range blocks {
		result, exists := svgs[block.Placeholder]
		if !exists {
			return "", nil, fmt.Errorf("missing Mermaid result for block %d", block.Index)
		}
		if result.Error != "" {
			return "", nil, fmt.Errorf("Mermaid block %d: %s", block.Index, result.Error)
		}
		file := assets.NewHashedFile(path.Join(section, slug, fmt.Sprintf("diagram-%d.svg", block.Index)), []byte(result.SVG))
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

func replaceParagraphPlaceholder(htmlFragment, placeholder, replacement string) (string, bool) {
	pattern := regexp.MustCompile(`<p>\s*` + regexp.QuoteMeta(placeholder) + `\s*</p>`)
	if !pattern.MatchString(htmlFragment) {
		return htmlFragment, false
	}
	return pattern.ReplaceAllString(htmlFragment, replacement), true
}
