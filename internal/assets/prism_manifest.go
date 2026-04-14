package assets

import (
	"fmt"
	"path"
	"sort"
	"strings"
)

// prismLanguageDependencies and prismLanguageAliases are manually maintained to
// keep build-time Prism selection independent from the minified autoloader JS.
var prismLanguageDependencies = map[string][]string{
	"actionscript":             {"javascript"},
	"apex":                     {"clike", "sql"},
	"arduino":                  {"cpp"},
	"aspnet":                   {"markup", "csharp"},
	"birb":                     {"clike"},
	"bison":                    {"c"},
	"c":                        {"clike"},
	"cfscript":                 {"clike"},
	"chaiscript":               {"clike", "cpp"},
	"cilkc":                    {"c"},
	"cilkcpp":                  {"cpp"},
	"coffeescript":             {"javascript"},
	"cpp":                      {"c"},
	"crystal":                  {"ruby"},
	"csharp":                   {"clike"},
	"cshtml":                   {"markup", "csharp"},
	"css-extras":               {"css"},
	"d":                        {"clike"},
	"dart":                     {"clike"},
	"django":                   {"markup-templating"},
	"ejs":                      {"javascript", "markup-templating"},
	"erb":                      {"ruby", "markup-templating"},
	"etlua":                    {"lua", "markup-templating"},
	"firestore-security-rules": {"clike"},
	"flow":                     {"javascript"},
	"fsharp":                   {"clike"},
	"ftl":                      {"markup-templating"},
	"glsl":                     {"c"},
	"gml":                      {"clike"},
	"go":                       {"clike"},
	"gradle":                   {"clike"},
	"groovy":                   {"clike"},
	"haml":                     {"ruby"},
	"handlebars":               {"markup-templating"},
	"haxe":                     {"clike"},
	"hlsl":                     {"c"},
	"idris":                    {"haskell"},
	"java":                     {"clike"},
	"javadoc":                  {"markup", "java", "javadoclike"},
	"javascript":               {"clike"},
	"jolie":                    {"clike"},
	"js-extras":                {"javascript"},
	"js-templates":             {"javascript"},
	"jsdoc":                    {"javascript", "javadoclike", "typescript"},
	"json5":                    {"json"},
	"jsonp":                    {"json"},
	"jsx":                      {"markup", "javascript"},
	"kotlin":                   {"clike"},
	"latte":                    {"clike", "markup-templating", "php"},
	"less":                     {"css"},
	"lilypond":                 {"scheme"},
	"liquid":                   {"markup-templating"},
	"markdown":                 {"markup"},
	"markup-templating":        {"markup"},
	"mongodb":                  {"javascript"},
	"n4js":                     {"javascript"},
	"objectivec":               {"c"},
	"opencl":                   {"c"},
	"parser":                   {"markup"},
	"php":                      {"markup-templating"},
	"php-extras":               {"php"},
	"phpdoc":                   {"php", "javadoclike"},
	"plsql":                    {"sql"},
	"processing":               {"clike"},
	"protobuf":                 {"clike"},
	"pug":                      {"markup", "javascript"},
	"purebasic":                {"clike"},
	"purescript":               {"haskell"},
	"qml":                      {"javascript"},
	"qore":                     {"clike"},
	"qsharp":                   {"clike"},
	"racket":                   {"scheme"},
	"reason":                   {"clike"},
	"ruby":                     {"clike"},
	"sass":                     {"css"},
	"scala":                    {"java"},
	"scss":                     {"css"},
	"shell-session":            {"bash"},
	"smarty":                   {"markup-templating"},
	"solidity":                 {"clike"},
	"soy":                      {"markup-templating"},
	"sparql":                   {"turtle"},
	"sqf":                      {"clike"},
	"squirrel":                 {"clike"},
	"stata":                    {"mata", "java", "python"},
	"t4-cs":                    {"t4-templating", "csharp"},
	"t4-vb":                    {"t4-templating", "vbnet"},
	"tap":                      {"yaml"},
	"textile":                  {"markup"},
	"tsx":                      {"jsx", "typescript"},
	"tt2":                      {"clike", "markup-templating"},
	"twig":                     {"markup-templating"},
	"typescript":               {"javascript"},
	"v":                        {"clike"},
	"vala":                     {"clike"},
	"vbnet":                    {"basic"},
	"velocity":                 {"markup"},
	"wiki":                     {"markup"},
	"xeora":                    {"markup"},
	"xml-doc":                  {"markup"},
	"xquery":                   {"markup"},
}

var prismLanguageAliases = map[string]string{
	"adoc":              "asciidoc",
	"arm-asm":           "armasm",
	"art":               "arturo",
	"atom":              "markup",
	"avdl":              "avro-idl",
	"avs":               "avisynth",
	"cfc":               "cfscript",
	"cilk":              "cilkcpp",
	"cilk-c":            "cilkc",
	"cilk-cpp":          "cilkcpp",
	"coffee":            "coffeescript",
	"conc":              "concurnas",
	"context":           "latex",
	"cs":                "csharp",
	"dns-zone":          "dns-zone-file",
	"dockerfile":        "docker",
	"dotnet":            "csharp",
	"elisp":             "lisp",
	"emacs":             "lisp",
	"emacs-lisp":        "lisp",
	"eta":               "ejs",
	"g4":                "antlr4",
	"gamemakerlanguage": "gml",
	"gawk":              "awk",
	"gitignore":         "ignore",
	"gni":               "gn",
	"go-mod":            "go-module",
	"gv":                "dot",
	"hbs":               "handlebars",
	"hgignore":          "ignore",
	"hs":                "haskell",
	"html":              "markup",
	"idr":               "idris",
	"ino":               "arduino",
	"jinja2":            "django",
	"js":                "javascript",
	"kt":                "kotlin",
	"kts":               "kotlin",
	"kum":               "kumir",
	"ld":                "linker-script",
	"ly":                "lilypond",
	"mathematica":       "wolfram",
	"mathml":            "markup",
	"md":                "markdown",
	"moon":              "moonscript",
	"mscript":           "powerquery",
	"mustache":          "handlebars",
	"n4jsd":             "n4js",
	"nani":              "naniscript",
	"nb":                "wolfram",
	"npmignore":         "ignore",
	"objc":              "objectivec",
	"objectpascal":      "pascal",
	"oscript":           "bsl",
	"pbfasm":            "purebasic",
	"pcode":             "peoplecode",
	"plantuml":          "plant-uml",
	"po":                "gettext",
	"pq":                "powerquery",
	"purs":              "purescript",
	"px":                "pcaxis",
	"py":                "python",
	"qasm":              "openqasm",
	"qs":                "qsharp",
	"razor":             "cshtml",
	"rb":                "ruby",
	"rbnf":              "bnf",
	"res":               "rescript",
	"rkt":               "racket",
	"robot":             "robotframework",
	"rpy":               "renpy",
	"rq":                "sparql",
	"rss":               "markup",
	"sclang":            "supercollider",
	"sh":                "bash",
	"sh-session":        "shell-session",
	"shell":             "bash",
	"shellsession":      "shell-session",
	"shortcode":         "bbcode",
	"sln":               "solution-file",
	"smlnj":             "sml",
	"sol":               "solidity",
	"ssml":              "markup",
	"svg":               "markup",
	"t4":                "t4-cs",
	"tex":               "latex",
	"trickle":           "tremor",
	"trig":              "turtle",
	"troy":              "tremor",
	"ts":                "typescript",
	"tsconfig":          "typoscript",
	"uc":                "unrealscript",
	"url":               "uri",
	"uscript":           "unrealscript",
	"vb":                "visual-basic",
	"vba":               "visual-basic",
	"webidl":            "web-idl",
	"webmanifest":       "json",
	"wl":                "wolfram",
	"xeoracube":         "xeora",
	"xls":               "excel-formula",
	"xlsx":              "excel-formula",
	"xml":               "markup",
	"yml":               "yaml",
}

var prismKnownLanguages = func() map[string]struct{} {
	languages := make(map[string]struct{}, len(prismLanguageDependencies)+len(prismLanguageAliases))
	for language, dependencies := range prismLanguageDependencies {
		languages[language] = struct{}{}
		for _, dependency := range dependencies {
			languages[dependency] = struct{}{}
		}
	}
	for _, language := range prismLanguageAliases {
		languages[language] = struct{}{}
	}
	return languages
}()

func normalizePrismLanguage(language string) string {
	canonical := strings.ToLower(strings.TrimSpace(language))
	if alias, ok := prismLanguageAliases[canonical]; ok {
		return alias
	}
	return canonical
}

func requiredPrismComponentLanguages(requested []string, available map[string][]byte) ([]string, error) {
	languages := append([]string(nil), requested...)
	sort.Strings(languages)

	selected := map[string]struct{}{}
	ordered := make([]string, 0, len(languages))
	var visit func(language string) error
	visit = func(language string) error {
		canonical := normalizePrismLanguage(language)
		if canonical == "" {
			return nil
		}
		if _, ok := selected[canonical]; ok {
			return nil
		}

		for _, dependency := range prismLanguageDependencies[canonical] {
			if err := visit(dependency); err != nil {
				return err
			}
		}

		componentPath := prismComponentPath(canonical)
		if _, ok := available[componentPath]; !ok {
			if _, ok := prismKnownLanguages[canonical]; ok {
				return fmt.Errorf("missing vendored Prism component %q", canonical)
			}
			return nil
		}

		selected[canonical] = struct{}{}
		ordered = append(ordered, canonical)
		return nil
	}

	for _, language := range languages {
		if err := visit(language); err != nil {
			return nil, err
		}
	}

	return ordered, nil
}

func prismComponentPath(language string) string {
	return path.Join(prismVersionDir, "components", "prism-"+language+".min.js")
}
