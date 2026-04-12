package content

import (
	"net/url"
	"path"
	"regexp"
	"strings"
)

var markdownLinkPattern = regexp.MustCompile(`!?\[[^\]]*\]\(([^)]+)\)`)

func CollectLocalAssetRefs(markdown string) []string {
	links := collectMarkdownLinks(markdown)
	results := make([]string, 0)
	seen := map[string]struct{}{}
	for _, link := range links {
		if strings.HasPrefix(link, "assets/") {
			if _, exists := seen[link]; !exists {
				seen[link] = struct{}{}
				results = append(results, link)
			}
		}
	}
	return results
}

func CollectInternalLinks(markdown string) []string {
	links := collectMarkdownLinks(markdown)
	results := make([]string, 0)
	seen := map[string]struct{}{}
	for _, link := range links {
		if isExternalLink(link) {
			continue
		}
		if strings.HasPrefix(link, "/") {
			if _, exists := seen[link]; !exists {
				seen[link] = struct{}{}
				results = append(results, link)
			}
		}
	}
	return results
}

func NormalizeURLPath(raw string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", err
	}
	if parsed.Path == "" {
		return "", nil
	}
	cleaned := path.Clean(parsed.Path)
	if cleaned == "." {
		cleaned = "/"
	}
	if strings.HasSuffix(parsed.Path, "/") && cleaned != "/" {
		cleaned += "/"
	}
	if cleaned != "/" && path.Ext(cleaned) == "" && !strings.HasSuffix(cleaned, "/") {
		cleaned += "/"
	}
	return cleaned, nil
}

func collectMarkdownLinks(markdown string) []string {
	lines := strings.Split(markdown, "\n")
	links := make([]string, 0)
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
				continue
			}
			matches := markdownLinkPattern.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				link := strings.Trim(strings.TrimSpace(match[1]), "<>")
				if link != "" {
					links = append(links, link)
				}
			}
			continue
		}
		if isFenceClose(line, fenceChar, fenceLen) {
			inFence = false
			fenceChar = 0
			fenceLen = 0
		}
	}

	return links
}

func isExternalLink(link string) bool {
	parsed, err := url.Parse(link)
	if err != nil {
		return false
	}
	return parsed.Scheme != "" || parsed.Host != "" || strings.HasPrefix(link, "mailto:")
}

func parseFence(line string) (bool, byte, int, string) {
	trimmed := strings.TrimLeft(line, " \t")
	if len(trimmed) < 3 {
		return false, 0, 0, ""
	}
	char := trimmed[0]
	if char != '`' && char != '~' {
		return false, 0, 0, ""
	}
	count := 0
	for count < len(trimmed) && trimmed[count] == char {
		count++
	}
	if count < 3 {
		return false, 0, 0, ""
	}
	return true, char, count, strings.TrimSpace(trimmed[count:])
}

func isFenceClose(line string, fenceChar byte, fenceLen int) bool {
	trimmed := strings.TrimLeft(line, " \t")
	if len(trimmed) < fenceLen {
		return false
	}
	for index := 0; index < fenceLen; index++ {
		if trimmed[index] != fenceChar {
			return false
		}
	}
	return strings.TrimSpace(trimmed[fenceLen:]) == ""
}
