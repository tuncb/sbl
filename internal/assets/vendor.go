package assets

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"sbl/embedded"
)

const (
	katexVersionDir   = "katex-0.16.45"
	mermaidVersionDir = "mermaid-11.14.0"
	prismVersionDir   = "prism-1.30.0"
	defaultPrismTheme = "prism"
)

type VendorAssets struct {
	KaTeXCSSURL          string
	KaTeXJSURL           string
	MermaidJSURL         string
	PrismCSSURL          string
	PrismCoreJSURL       string
	PrismAutoloaderJSURL string
	PrismLanguagesPath   string
}

type VendorRequest struct {
	IncludeKaTeX   bool
	IncludeMermaid bool
	PrismLanguages []string
	PrismTheme     string
}

var prismThemeFiles = map[string]string{
	"prism":          path.Join(prismVersionDir, "themes", "prism.min.css"),
	"coy":            path.Join(prismVersionDir, "themes", "prism-coy.min.css"),
	"dark":           path.Join(prismVersionDir, "themes", "prism-dark.min.css"),
	"funky":          path.Join(prismVersionDir, "themes", "prism-funky.min.css"),
	"okaidia":        path.Join(prismVersionDir, "themes", "prism-okaidia.min.css"),
	"solarizedlight": path.Join(prismVersionDir, "themes", "prism-solarizedlight.min.css"),
	"tomorrow":       path.Join(prismVersionDir, "themes", "prism-tomorrow.min.css"),
	"twilight":       path.Join(prismVersionDir, "themes", "prism-twilight.min.css"),
}

func DefaultPrismTheme() string {
	return defaultPrismTheme
}

func SupportedPrismThemes() []string {
	themes := make([]string, 0, len(prismThemeFiles))
	for theme := range prismThemeFiles {
		themes = append(themes, theme)
	}
	sort.Strings(themes)
	return themes
}

func NormalizePrismTheme(theme string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(theme))
	if normalized == "" {
		normalized = defaultPrismTheme
	}
	if _, ok := prismThemeFiles[normalized]; !ok {
		return "", fmt.Errorf("unknown prism_theme %q; supported themes: %s", theme, strings.Join(SupportedPrismThemes(), ", "))
	}
	return normalized, nil
}

func DefaultVendorAssets(prismTheme string) (VendorAssets, error) {
	theme, err := NormalizePrismTheme(prismTheme)
	if err != nil {
		return VendorAssets{}, err
	}
	prismThemeFile := prismThemeFiles[theme]
	return VendorAssets{
		KaTeXCSSURL:          "/" + path.Join("assets", "vendor", katexVersionDir, "katex.min.css"),
		KaTeXJSURL:           "/" + path.Join("assets", "vendor", katexVersionDir, "katex.min.js"),
		MermaidJSURL:         "/" + path.Join("assets", "vendor", mermaidVersionDir, "mermaid.min.js"),
		PrismCSSURL:          "/" + path.Join("assets", "vendor", prismThemeFile),
		PrismCoreJSURL:       "/" + path.Join("assets", "vendor", prismVersionDir, "components", "prism-core.min.js"),
		PrismAutoloaderJSURL: "/" + path.Join("assets", "vendor", prismVersionDir, "plugins", "autoloader", "prism-autoloader.min.js"),
		PrismLanguagesPath:   "/" + path.Join("assets", "vendor", prismVersionDir, "components") + "/",
	}, nil
}

func BuildVendorFiles(request VendorRequest) ([]File, VendorAssets, error) {
	vendorAssets, err := DefaultVendorAssets(request.PrismTheme)
	if err != nil {
		return nil, VendorAssets{}, err
	}

	files := map[string][]byte{}
	if err := readFSFiles(embedded.Vendor, files); err != nil {
		return nil, VendorAssets{}, err
	}

	selected := map[string]struct{}{}
	addFamily := func(prefix string) {
		for rel := range files {
			if rel == prefix || strings.HasPrefix(rel, prefix+"/") {
				selected[rel] = struct{}{}
			}
		}
	}
	requireFile := func(rel, label string) error {
		if _, ok := files[rel]; !ok {
			return fmt.Errorf("missing vendored %s", label)
		}
		selected[rel] = struct{}{}
		return nil
	}

	if request.IncludeKaTeX {
		addFamily(katexVersionDir)
		if err := requireFile(path.Join(katexVersionDir, "katex.min.css"), "KaTeX stylesheet"); err != nil {
			return nil, VendorAssets{}, err
		}
		if err := requireFile(path.Join(katexVersionDir, "katex.min.js"), "KaTeX script"); err != nil {
			return nil, VendorAssets{}, err
		}
	}
	if request.IncludeMermaid {
		addFamily(mermaidVersionDir)
		if err := requireFile(path.Join(mermaidVersionDir, "mermaid.min.js"), "Mermaid script"); err != nil {
			return nil, VendorAssets{}, err
		}
	}
	if len(request.PrismLanguages) > 0 {
		theme, err := NormalizePrismTheme(request.PrismTheme)
		if err != nil {
			return nil, VendorAssets{}, err
		}
		if err := requireFile(prismThemeFiles[theme], "Prism stylesheet"); err != nil {
			return nil, VendorAssets{}, err
		}
		if err := requireFile(path.Join(prismVersionDir, "components", "prism-core.min.js"), "Prism core script"); err != nil {
			return nil, VendorAssets{}, err
		}
		if err := requireFile(path.Join(prismVersionDir, "plugins", "autoloader", "prism-autoloader.min.js"), "Prism autoloader script"); err != nil {
			return nil, VendorAssets{}, err
		}
		if err := requireFile(path.Join(prismVersionDir, "LICENSE"), "Prism license"); err != nil {
			return nil, VendorAssets{}, err
		}
		languages, err := requiredPrismComponentLanguages(request.PrismLanguages, files)
		if err != nil {
			return nil, VendorAssets{}, err
		}
		for _, language := range languages {
			if err := requireFile(prismComponentPath(language), "Prism component "+language); err != nil {
				return nil, VendorAssets{}, err
			}
		}
	}

	paths := make([]string, 0, len(selected))
	for rel := range selected {
		paths = append(paths, rel)
	}
	sort.Strings(paths)

	out := make([]File, 0, len(paths))
	for _, rel := range paths {
		cleaned := path.Clean(rel)
		target := path.Join("assets", "vendor", cleaned)
		out = append(out, File{
			RelPath: target,
			URL:     "/" + target,
			Bytes:   files[rel],
		})
	}

	return out, vendorAssets, nil
}
