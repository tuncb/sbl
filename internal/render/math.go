package render

import (
	"bytes"
	"fmt"
	"html"
	"strings"

	nethtml "golang.org/x/net/html"
	"golang.org/x/net/html/atom"
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
		replacement := fmt.Sprintf(`<div class="math math-display"><code>%s</code></div>`, html.EscapeString(block.Source))
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
		replacement := fmt.Sprintf(`<span class="math math-inline"><code>%s</code></span>`, html.EscapeString(block.Source))
		if !strings.Contains(htmlFragment, block.Placeholder) {
			return "", fmt.Errorf("missing inline math placeholder %q in rendered HTML", block.Placeholder)
		}
		htmlFragment = strings.ReplaceAll(htmlFragment, block.Placeholder, replacement)
	}
	return htmlFragment, nil
}

func RenderInlineMath(htmlFragment string) (string, error) {
	root, err := parseHTMLFragment(htmlFragment)
	if err != nil {
		return "", err
	}
	if err := walkTextNodes(root, false); err != nil {
		return "", err
	}
	return renderHTMLFragment(root)
}

func walkTextNodes(node *nethtml.Node, skip bool) error {
	nextSkip := skip || isCodeLikeNode(node)
	for child := node.FirstChild; child != nil; {
		next := child.NextSibling
		if child.Type == nethtml.TextNode && !nextSkip {
			replacements, changed, err := renderInlineMathText(child.Data)
			if err != nil {
				return err
			}
			if changed {
				parent := child.Parent
				for _, replacement := range replacements {
					parent.InsertBefore(replacement, child)
				}
				parent.RemoveChild(child)
			}
		} else if err := walkTextNodes(child, nextSkip); err != nil {
			return err
		}
		child = next
	}
	return nil
}

func renderInlineMathText(text string) ([]*nethtml.Node, bool, error) {
	if !strings.Contains(text, `\(`) {
		if strings.Contains(text, `\)`) || strings.Contains(text, "$$") {
			return nil, false, fmt.Errorf("invalid math delimiters in %q", text)
		}
		return nil, false, nil
	}

	nodes := make([]*nethtml.Node, 0)
	remaining := text
	for {
		start := strings.Index(remaining, `\(`)
		if start < 0 {
			if strings.Contains(remaining, `\)`) || strings.Contains(remaining, "$$") {
				return nil, false, fmt.Errorf("invalid math delimiters in %q", text)
			}
			if remaining != "" {
				nodes = append(nodes, &nethtml.Node{Type: nethtml.TextNode, Data: remaining})
			}
			break
		}
		if start > 0 {
			nodes = append(nodes, &nethtml.Node{Type: nethtml.TextNode, Data: remaining[:start]})
		}
		rest := remaining[start+2:]
		end := strings.Index(rest, `\)`)
		if end < 0 {
			return nil, false, fmt.Errorf("unterminated inline math delimiter in %q", text)
		}
		expr := strings.TrimSpace(rest[:end])
		if expr == "" {
			return nil, false, fmt.Errorf("empty inline math expression in %q", text)
		}
		nodes = append(nodes, mathNode("span", "math math-inline", expr))
		remaining = rest[end+2:]
	}

	return nodes, true, nil
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

func mathNode(tag, className, expression string) *nethtml.Node {
	element := &nethtml.Node{Type: nethtml.ElementNode, Data: tag}
	switch tag {
	case "span":
		element.DataAtom = atom.Span
	case "div":
		element.DataAtom = atom.Div
	}
	element.Attr = append(element.Attr, nethtml.Attribute{Key: "class", Val: className})

	code := &nethtml.Node{Type: nethtml.ElementNode, Data: "code", DataAtom: atom.Code}
	code.AppendChild(&nethtml.Node{Type: nethtml.TextNode, Data: expression})
	element.AppendChild(code)
	return element
}

func RewriteAssetLinks(htmlFragment string, assetURLs map[string]string) (string, error) {
	root, err := parseHTMLFragment(htmlFragment)
	if err != nil {
		return "", err
	}
	rewriteAssetNodes(root, assetURLs)
	return renderHTMLFragment(root)
}

func rewriteAssetNodes(node *nethtml.Node, assetURLs map[string]string) {
	if node.Type == nethtml.ElementNode {
		for index := range node.Attr {
			if node.Attr[index].Key != "src" && node.Attr[index].Key != "href" {
				continue
			}
			if target, exists := assetURLs[node.Attr[index].Val]; exists {
				node.Attr[index].Val = target
			}
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		rewriteAssetNodes(child, assetURLs)
	}
}

func parseHTMLFragment(fragment string) (*nethtml.Node, error) {
	container := &nethtml.Node{Type: nethtml.ElementNode, Data: "div", DataAtom: atom.Div}
	nodes, err := nethtml.ParseFragment(strings.NewReader(fragment), container)
	if err != nil {
		return nil, err
	}
	root := &nethtml.Node{Type: nethtml.ElementNode, Data: "div", DataAtom: atom.Div}
	for _, node := range nodes {
		root.AppendChild(node)
	}
	return root, nil
}

func renderHTMLFragment(root *nethtml.Node) (string, error) {
	var buffer bytes.Buffer
	for child := root.FirstChild; child != nil; child = child.NextSibling {
		if err := nethtml.Render(&buffer, child); err != nil {
			return "", err
		}
	}
	return buffer.String(), nil
}

func isCodeLikeNode(node *nethtml.Node) bool {
	if node.Type != nethtml.ElementNode {
		return false
	}
	return node.DataAtom == atom.Pre || node.DataAtom == atom.Code
}
