package render

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"sbl/embedded"
	"sbl/internal/content"
	"sbl/internal/site"
)

type Engine struct {
	files map[string]string
}

type BasePageData struct {
	Site             site.Config
	Title            string
	Description      string
	CanonicalURL     string
	StylesheetURL    string
	ExtraStylesheets []string
}

type PostSummary struct {
	Title   string
	Summary string
	URL     string
	Date    string
}

type ListPageData struct {
	BasePageData
	Posts []PostSummary
}

type PostView struct {
	Title       string
	Summary     string
	BodyHTML    template.HTML
	Date        string
	Updated     string
	ReadingTime int
}

type PostPageData struct {
	BasePageData
	Post PostView
}

type PageView struct {
	Title    string
	Summary  string
	BodyHTML template.HTML
}

type StandalonePageData struct {
	BasePageData
	Page PageView
}

func NewTemplateEngine(siteRoot string) (*Engine, error) {
	files := map[string]string{}
	if err := loadTemplateFS(embedded.Templates, files); err != nil {
		return nil, err
	}

	overrideDir := filepath.Join(siteRoot, "templates")
	if err := loadTemplateDir(overrideDir, files); err != nil {
		return nil, err
	}

	return &Engine{files: files}, nil
}

func (e *Engine) Execute(page string, data any) ([]byte, error) {
	base, ok := e.files["base.html"]
	if !ok {
		return nil, fmt.Errorf("missing base.html template")
	}
	pageTemplate, ok := e.files[page]
	if !ok {
		return nil, fmt.Errorf("missing %s template", page)
	}

	tmpl := template.New("page")
	var err error
	if tmpl, err = tmpl.Parse(base); err != nil {
		return nil, err
	}

	partialNames := make([]string, 0)
	for name := range e.files {
		if strings.HasPrefix(name, "partials/") {
			partialNames = append(partialNames, name)
		}
	}
	sort.Strings(partialNames)
	for _, name := range partialNames {
		if tmpl, err = tmpl.Parse(e.files[name]); err != nil {
			return nil, err
		}
	}

	if tmpl, err = tmpl.Parse(pageTemplate); err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buffer, "page", data); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func MakePostSummary(post *content.Post) PostSummary {
	return PostSummary{
		Title:   post.Title,
		Summary: post.Summary,
		URL:     post.CanonicalPath,
		Date:    post.Date.Format("2006-01-02"),
	}
}

func RenderIndexPage(engine *Engine, cfg site.Config, stylesheetURL string, extraStylesheets []string, posts []PostSummary) ([]byte, error) {
	return engine.Execute("index.html", ListPageData{
		BasePageData: BasePageData{
			Site:             cfg,
			Title:            cfg.Title,
			Description:      cfg.Description,
			CanonicalURL:     cfg.CanonicalURL("/"),
			StylesheetURL:    stylesheetURL,
			ExtraStylesheets: extraStylesheets,
		},
		Posts: posts,
	})
}

func RenderArchivePage(engine *Engine, cfg site.Config, stylesheetURL string, extraStylesheets []string, posts []PostSummary) ([]byte, error) {
	return engine.Execute("archive.html", ListPageData{
		BasePageData: BasePageData{
			Site:             cfg,
			Title:            "Archive | " + cfg.Title,
			Description:      "Archive of posts from " + cfg.Title,
			CanonicalURL:     cfg.CanonicalURL("/archive/"),
			StylesheetURL:    stylesheetURL,
			ExtraStylesheets: extraStylesheets,
		},
		Posts: posts,
	})
}

func RenderPostPage(engine *Engine, cfg site.Config, stylesheetURL string, extraStylesheets []string, post *content.Post, bodyHTML template.HTML, readingTime int) ([]byte, PostSummary, error) {
	htmlPage, err := engine.Execute("post.html", PostPageData{
		BasePageData: BasePageData{
			Site:             cfg,
			Title:            post.Title + " | " + cfg.Title,
			Description:      pickDescription(post),
			CanonicalURL:     cfg.CanonicalURL(post.CanonicalPath),
			StylesheetURL:    stylesheetURL,
			ExtraStylesheets: extraStylesheets,
		},
		Post: PostView{
			Title:       post.Title,
			Summary:     post.Summary,
			BodyHTML:    bodyHTML,
			Date:        post.Date.Format("2006-01-02"),
			Updated:     formatOptionalDate(post.Updated),
			ReadingTime: readingTime,
		},
	})
	if err != nil {
		return nil, PostSummary{}, err
	}
	return htmlPage, MakePostSummary(post), nil
}

func RenderStandalonePage(engine *Engine, cfg site.Config, stylesheetURL string, extraStylesheets []string, page *content.Page, bodyHTML template.HTML) ([]byte, error) {
	return engine.Execute("page.html", StandalonePageData{
		BasePageData: BasePageData{
			Site:             cfg,
			Title:            page.Title + " | " + cfg.Title,
			Description:      pickPageDescription(page),
			CanonicalURL:     cfg.CanonicalURL(page.CanonicalPath),
			StylesheetURL:    stylesheetURL,
			ExtraStylesheets: extraStylesheets,
		},
		Page: PageView{
			Title:    page.Title,
			Summary:  page.Summary,
			BodyHTML: bodyHTML,
		},
	})
}

func RenderNotFoundPage(engine *Engine, cfg site.Config, stylesheetURL string, extraStylesheets []string) ([]byte, error) {
	return engine.Execute("404.html", BasePageData{
		Site:             cfg,
		Title:            "Page not found",
		Description:      "The page you asked for does not exist or has moved.",
		CanonicalURL:     cfg.CanonicalURL("/404.html"),
		StylesheetURL:    stylesheetURL,
		ExtraStylesheets: extraStylesheets,
	})
}

func RenderTemporaryErrorPage(engine *Engine, cfg site.Config, stylesheetURL string, extraStylesheets []string) ([]byte, error) {
	return engine.Execute("50x.html", BasePageData{
		Site:             cfg,
		Title:            "Temporary server problem",
		Description:      "Please try again in a moment.",
		CanonicalURL:     cfg.CanonicalURL("/50x.html"),
		StylesheetURL:    stylesheetURL,
		ExtraStylesheets: extraStylesheets,
	})
}

func loadTemplateFS(source fs.FS, out map[string]string) error {
	return fs.WalkDir(source, ".", func(rel string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		data, err := fs.ReadFile(source, rel)
		if err != nil {
			return err
		}
		out[filepath.ToSlash(rel)] = string(data)
		return nil
	})
}

func loadTemplateDir(root string, out map[string]string) error {
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	return filepath.WalkDir(root, func(filePath string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		data, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, filePath)
		if err != nil {
			return err
		}
		out[filepath.ToSlash(rel)] = string(data)
		return nil
	})
}

func pickDescription(post *content.Post) string {
	if strings.TrimSpace(post.Description) != "" {
		return post.Description
	}
	return post.Summary
}

func pickPageDescription(page *content.Page) string {
	if strings.TrimSpace(page.Description) != "" {
		return page.Description
	}
	return page.Summary
}

func formatOptionalDate(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format("2006-01-02")
}
