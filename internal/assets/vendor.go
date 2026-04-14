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
}

func DefaultVendorAssets() VendorAssets {
	return VendorAssets{
		KaTeXCSSURL:          "/" + path.Join("assets", "vendor", katexVersionDir, "katex.min.css"),
		KaTeXJSURL:           "/" + path.Join("assets", "vendor", katexVersionDir, "katex.min.js"),
		MermaidJSURL:         "/" + path.Join("assets", "vendor", mermaidVersionDir, "mermaid.min.js"),
		PrismCSSURL:          "/" + path.Join("assets", "vendor", prismVersionDir, "themes", "prism.min.css"),
		PrismCoreJSURL:       "/" + path.Join("assets", "vendor", prismVersionDir, "components", "prism-core.min.js"),
		PrismAutoloaderJSURL: "/" + path.Join("assets", "vendor", prismVersionDir, "plugins", "autoloader", "prism-autoloader.min.js"),
		PrismLanguagesPath:   "/" + path.Join("assets", "vendor", prismVersionDir, "components") + "/",
	}
}

func BuildVendorFiles(request VendorRequest) ([]File, VendorAssets, error) {
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
		if err := requireFile(path.Join(prismVersionDir, "themes", "prism.min.css"), "Prism stylesheet"); err != nil {
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

	return out, DefaultVendorAssets(), nil
}
