package assets

import (
	"fmt"
	"path"
	"sort"

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

func BuildVendorFiles() ([]File, VendorAssets, error) {
	files := map[string][]byte{}
	if err := readFSFiles(embedded.Vendor, files); err != nil {
		return nil, VendorAssets{}, err
	}

	paths := make([]string, 0, len(files))
	for rel := range files {
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

	assets := VendorAssets{
		KaTeXCSSURL:          "/" + path.Join("assets", "vendor", katexVersionDir, "katex.min.css"),
		KaTeXJSURL:           "/" + path.Join("assets", "vendor", katexVersionDir, "katex.min.js"),
		MermaidJSURL:         "/" + path.Join("assets", "vendor", mermaidVersionDir, "mermaid.min.js"),
		PrismCSSURL:          "/" + path.Join("assets", "vendor", prismVersionDir, "themes", "prism.min.css"),
		PrismCoreJSURL:       "/" + path.Join("assets", "vendor", prismVersionDir, "components", "prism-core.min.js"),
		PrismAutoloaderJSURL: "/" + path.Join("assets", "vendor", prismVersionDir, "plugins", "autoloader", "prism-autoloader.min.js"),
		PrismLanguagesPath:   "/" + path.Join("assets", "vendor", prismVersionDir, "components") + "/",
	}
	if _, ok := files[path.Join(katexVersionDir, "katex.min.css")]; !ok {
		return nil, VendorAssets{}, fmt.Errorf("missing vendored KaTeX stylesheet")
	}
	if _, ok := files[path.Join(katexVersionDir, "katex.min.js")]; !ok {
		return nil, VendorAssets{}, fmt.Errorf("missing vendored KaTeX script")
	}
	if _, ok := files[path.Join(mermaidVersionDir, "mermaid.min.js")]; !ok {
		return nil, VendorAssets{}, fmt.Errorf("missing vendored Mermaid script")
	}
	if _, ok := files[path.Join(prismVersionDir, "themes", "prism.min.css")]; !ok {
		return nil, VendorAssets{}, fmt.Errorf("missing vendored Prism stylesheet")
	}
	if _, ok := files[path.Join(prismVersionDir, "components", "prism-core.min.js")]; !ok {
		return nil, VendorAssets{}, fmt.Errorf("missing vendored Prism core script")
	}
	if _, ok := files[path.Join(prismVersionDir, "plugins", "autoloader", "prism-autoloader.min.js")]; !ok {
		return nil, VendorAssets{}, fmt.Errorf("missing vendored Prism autoloader script")
	}

	return out, assets, nil
}
