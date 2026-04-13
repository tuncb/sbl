package render

import (
	"fmt"
	"html"
	"strings"
)

type DisplayMathBlock struct {
	Index       int
	Placeholder string
	Source      string
}

type InlineMathBlock struct {
	Index       int
	Placeholder string
	Source      string
}

func ExtractDisplayMath(markdown string) (string, []DisplayMathBlock, error) {
	lines := strings.Split(markdown, "\n")
	var output []string
	blocks := make([]DisplayMathBlock, 0)
	var fenceChar byte
	var fenceLen int
	inFence := false
	inDisplay := false
	var current []string

	for _, rawLine := range lines {
		line := strings.TrimRight(rawLine, "\r")
		trimmed := strings.TrimSpace(line)

		if !inFence && !inDisplay {
			if matched, char, count, _ := parseFence(line); matched {
				inFence = true
				fenceChar = char
				fenceLen = count
				output = append(output, line)
				continue
			}
			if strings.HasPrefix(trimmed, "$$") && strings.HasSuffix(trimmed, "$$") && len(trimmed) > 4 {
				index := len(blocks) + 1
				placeholder := fmt.Sprintf("SBL_DISPLAY_MATH_%d", index)
				source := strings.TrimSpace(trimmed[2 : len(trimmed)-2])
				if source == "" {
					return "", nil, fmt.Errorf("display math block %d is empty", index)
				}
				blocks = append(blocks, DisplayMathBlock{Index: index, Placeholder: placeholder, Source: source})
				output = append(output, placeholder)
				continue
			}
			if trimmed == "$$" {
				inDisplay = true
				current = current[:0]
				continue
			}
			output = append(output, line)
			continue
		}

		if inFence {
			output = append(output, line)
			if isFenceClose(line, fenceChar, fenceLen) {
				inFence = false
				fenceChar = 0
				fenceLen = 0
			}
			continue
		}

		if trimmed == "$$" {
			index := len(blocks) + 1
			source := strings.TrimSpace(strings.Join(current, "\n"))
			if source == "" {
				return "", nil, fmt.Errorf("display math block %d is empty", index)
			}
			placeholder := fmt.Sprintf("SBL_DISPLAY_MATH_%d", index)
			blocks = append(blocks, DisplayMathBlock{Index: index, Placeholder: placeholder, Source: source})
			output = append(output, placeholder)
			inDisplay = false
			current = current[:0]
			continue
		}
		current = append(current, line)
	}

	if inDisplay {
		return "", nil, fmt.Errorf("unterminated display math block")
	}

	return strings.Join(output, "\n"), blocks, nil
}

func ReplaceDisplayMathPlaceholders(htmlFragment string, blocks []DisplayMathBlock) (string, error) {
	for _, block := range blocks {
		replacement := `<div class="sbl-math-display">` + html.EscapeString(block.Source) + `</div>`
		var replaced bool
		htmlFragment, replaced = replaceParagraphPlaceholder(htmlFragment, block.Placeholder, replacement)
		if !replaced {
			return "", fmt.Errorf("missing display math placeholder %q in rendered HTML", block.Placeholder)
		}
	}
	return htmlFragment, nil
}

func ExtractInlineMath(markdown string) (string, []InlineMathBlock, error) {
	lines := strings.Split(markdown, "\n")
	blocks := make([]InlineMathBlock, 0)
	var output []string
	var fenceChar byte
	var fenceLen int
	inFence := false

	for _, rawLine := range lines {
		line := strings.TrimRight(rawLine, "\r")
		if !inFence {
			if matched, char, count, _ := parseFence(line); matched {
				inFence = true
				fenceChar = char
				fenceLen = count
				output = append(output, line)
				continue
			}
			replaced, newBlocks, err := extractInlineMathLine(line, len(blocks))
			if err != nil {
				return "", nil, err
			}
			blocks = append(blocks, newBlocks...)
			output = append(output, replaced)
			continue
		}
		output = append(output, line)
		if isFenceClose(line, fenceChar, fenceLen) {
			inFence = false
			fenceChar = 0
			fenceLen = 0
		}
	}

	return strings.Join(output, "\n"), blocks, nil
}

func ReplaceInlineMathPlaceholders(htmlFragment string, blocks []InlineMathBlock) (string, error) {
	for _, block := range blocks {
		if !strings.Contains(htmlFragment, block.Placeholder) {
			return "", fmt.Errorf("missing inline math placeholder %q in rendered HTML", block.Placeholder)
		}
		replacement := `<span class="sbl-math-inline">` + html.EscapeString(block.Source) + `</span>`
		htmlFragment = strings.ReplaceAll(htmlFragment, block.Placeholder, replacement)
	}
	return htmlFragment, nil
}

func extractInlineMathLine(line string, offset int) (string, []InlineMathBlock, error) {
	remaining := line
	var builder strings.Builder
	blocks := make([]InlineMathBlock, 0)

	for {
		start := strings.Index(remaining, `\(`)
		if start < 0 {
			if strings.Contains(remaining, `\)`) {
				return "", nil, fmt.Errorf("invalid inline math delimiters in %q", line)
			}
			builder.WriteString(remaining)
			break
		}

		builder.WriteString(remaining[:start])
		rest := remaining[start+2:]
		end := strings.Index(rest, `\)`)
		if end < 0 {
			return "", nil, fmt.Errorf("unterminated inline math delimiter in %q", line)
		}
		expr := strings.TrimSpace(rest[:end])
		if expr == "" {
			return "", nil, fmt.Errorf("empty inline math expression in %q", line)
		}
		index := offset + len(blocks) + 1
		placeholder := fmt.Sprintf("SBL_INLINE_MATH_%d", index)
		blocks = append(blocks, InlineMathBlock{
			Index:       index,
			Placeholder: placeholder,
			Source:      expr,
		})
		builder.WriteString(placeholder)
		remaining = rest[end+2:]
	}

	return builder.String(), blocks, nil
}
