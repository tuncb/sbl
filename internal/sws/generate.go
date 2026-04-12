package sws

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func Generate(siteRoot, outputDir string, aliasRedirects map[string]string) (string, error) {
	base, err := loadBase(siteRoot)
	if err != nil {
		return "", err
	}
	base = patchRoot(base, siteRoot, outputDir)

	keys := make([]string, 0, len(aliasRedirects))
	for source := range aliasRedirects {
		keys = append(keys, source)
	}
	sort.Strings(keys)

	var builder strings.Builder
	builder.WriteString(base)
	if !strings.HasSuffix(base, "\n") {
		builder.WriteString("\n")
	}
	for _, source := range keys {
		builder.WriteString("\n[[advanced.redirects]]\n")
		builder.WriteString(fmt.Sprintf("source = %q\n", source))
		builder.WriteString(fmt.Sprintf("destination = %q\n", aliasRedirects[source]))
		builder.WriteString("kind = 301\n")
	}
	return builder.String(), nil
}

func Write(siteRoot, outputDir string, aliasRedirects map[string]string) error {
	content, err := Generate(siteRoot, outputDir, aliasRedirects)
	if err != nil {
		return err
	}
	deployDir := filepath.Join(siteRoot, "deploy")
	if err := os.MkdirAll(deployDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(deployDir, "sws.toml"), []byte(content), 0o644)
}

func patchRoot(config, siteRoot, outputDir string) string {
	rel, err := filepath.Rel(siteRoot, outputDir)
	if err != nil {
		return config
	}
	rel = filepath.ToSlash(rel)
	if rel == "public" {
		return config
	}
	if !strings.HasPrefix(rel, ".") && !strings.HasPrefix(rel, "/") {
		rel = "./" + rel
	}
	lines := strings.Split(config, "\n")
	for index, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "root = ") {
			lines[index] = fmt.Sprintf("root = %q", rel)
			break
		}
	}
	return strings.Join(lines, "\n")
}
